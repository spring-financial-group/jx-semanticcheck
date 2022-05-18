package check

import (
	"fmt"
	chgit "github.com/antham/chyle/chyle/git"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"strings"
)

// Options contains the command line flags
type Options struct {
	options.BaseOptions

	ScmFactory    scmhelpers.Options
	GitClient     gitclient.Interface
	CommandRunner cmdrunner.CommandRunner

	firstSha  string
	latestSha string
	dir       string
}

var (
	cmdLong = templates.LongDesc(`
		Checks whether the commit messages in a pull request follow the Conventional Commits specification
`)

	cmdExample = templates.Examples(`
		jx-semanticcheck check 
`)

	ConventionalCommitTypes = []string{"feat", "fix", "perf", "refactor", "docs", "test", "revert", "style", "chore", "build"}
)

// NewCmdCheckSemantics creates a command object for the command
func NewCmdCheckSemantics() (*cobra.Command, *Options) {
	o := &Options{}
	cmd := &cobra.Command{
		Use:     "check",
		Short:   "Checks for whether the commits in a PR are Conventional Commits",
		Long:    cmdLong,
		Example: cmdExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.ScmFactory.DiscoverFromGit = true
	cmd.Flags().StringVarP(&o.firstSha, "first-sha", "", "", "the first commit SHA to check")
	cmd.Flags().StringVarP(&o.latestSha, "latest-sha", "", "", "the last commit SHA to check")
	cmd.Flags().StringVarP(&o.dir, "dir", "", "", "the directory of the repository")

	return cmd, o
}

func (o *Options) Run() error {
	err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate")
	}

	commits, err := o.GetCurrentRevCommits()
	if err != nil {
		return errors.Wrapf(err, "failed to get commits")
	}

	if strings.HasPrefix((*commits)[0].Message, "release ") {
		// Ignore release commit
		*commits = (*commits)[1:]
	}

	var failedCommits int
	for _, commit := range *commits {
		var terminalMessage string
		indicator := "✓"

		if !IsCommitConventional(commit.Message) {
			indicator = "x"
			terminalMessage = commit.Message
			failedCommits++
		}

		log.Logger().Infof("---  Commit | %s --- %s\n"+
			"%s",
			commit.Hash, indicator, terminalMessage)
	}

	if failedCommits > 0 {
		return fmt.Errorf("%d commit(s) did not follow https://conventionalcommits.org/", failedCommits)
	}

	log.Logger().Infof("\nAll commits follow https://conventionalcommits.org/")
	return nil
}

// Validate checks that all the variables required to run are present
func (o *Options) Validate() error {
	err := o.BaseOptions.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate base options")
	}

	err = o.ScmFactory.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to discover git repository")
	}

	if o.dir == "" {
		o.dir = o.ScmFactory.Dir
	}

	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.CommandRunner)
	}
	return nil
}

// GetCurrentRevCommits returns the commits since the last revision
func (o *Options) GetCurrentRevCommits() (commits *[]object.Commit, err error) {
	if o.firstSha == "" {
		o.firstSha, _, err = gitclient.GetCommitPointedToByLatestTag(o.GitClient, o.dir)
		if err != nil {
			return nil, err
		}
	}

	if o.latestSha == "" {
		o.latestSha, err = gitclient.GetLatestCommitSha(o.GitClient, o.dir)
		if err != nil {
			return nil, err
		}
	}
	return chgit.FetchCommits(o.dir, o.firstSha, o.latestSha)
}

// IsCommitConventional checks whether a commit message follows the conventions by comparing its prefix
// to those in ConventionalCommitTypes
func IsCommitConventional(commitMessage string) bool {
	commitMessage = strings.TrimSpace(strings.ToLower(commitMessage))

	// Ignore revert or merge commits
	if strings.Contains(commitMessage, "revert") || strings.Contains(commitMessage, "merge") {
		return true
	}

	idx := strings.Index(commitMessage, ":")
	if idx > 0 {
		commitType := commitMessage[0:idx]
		for _, conventionalType := range ConventionalCommitTypes {
			if strings.HasPrefix(commitType, conventionalType) {
				return true
			}
		}
	}
	return false
}
