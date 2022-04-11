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
	"strings"
)

// Options contains the command line flags
type Options struct {
	options.BaseOptions

	ScmFactory    scmhelpers.Options
	GitClient     gitclient.Interface
	CommandRunner cmdrunner.CommandRunner

	Namespace       string
	LatestCommitSha string
	CurrentRevSha   string
	RepoDir         string
}

var (
	cmdLong = templates.LongDesc(`
		Checks whether the commit messages in a pull request follow Conventional Commits.
`)

	cmdExample = templates.Examples(`
		jx-semanticcheck check 
`)

	ConventionalCommitTitles = []string{"feat", "fix", "perf", "refactor", "docs", "test", "revert", "style", "chore"}
)

func NewCmdCheckSemantics() (*cobra.Command, *Options) {
	o := &Options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Checks for whether the commits are conventional",
		// Aliases: []string{"changelog", "changes", "publish"},
		Long:    cmdLong,
		Example: cmdExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := o.Run()
			helper.CheckErr(err)
		},
	}
	o.ScmFactory.DiscoverFromGit = true
	cmd.Flags().StringVarP(&o.CurrentRevSha, "first-sha", "p", "", "the first commit SHA of this revision")
	cmd.Flags().StringVarP(&o.LatestCommitSha, "latest-sha", "", "", "the latest commit SHA")
	cmd.Flags().StringVarP(&o.RepoDir, "repo-dir", "", "", "the directory of the git repository")

	return cmd, o
}

func (o *Options) Validate() error {
	err := o.BaseOptions.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate base options")
	}

	err = o.ScmFactory.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to discover git repository")
	}

	return nil
}

func (o *Options) Run() error {
	err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate")
	}

	o.CommandRunner = cmdrunner.QuietCommandRunner

	dir := o.RepoDir
	if dir == "" {
		dir = o.ScmFactory.Dir
	}

	if o.CurrentRevSha == "" {
		o.CurrentRevSha, _, err = gitclient.GetCommitPointedToByLatestTag(o.git(), dir)
		if err != nil {
			return err
		}
	}

	if o.LatestCommitSha == "" {
		o.LatestCommitSha, err = gitclient.GetLatestCommitSha(o.git(), dir)
		if err != nil {
			return err
		}
	}

	commits, err := chgit.FetchCommits(dir, o.CurrentRevSha, o.LatestCommitSha)
	if err != nil {
		return err
	}

	commitSlice := *commits
	if strings.HasPrefix(commitSlice[0].Message, "release ") {
		// remove the release commit from the log
		commitSlice = commitSlice[1:]
	}

	var failedCommitCount int
	for _, commit := range commitSlice {
		var terminalMessage string
		passMark := "✓"

		if !IsCommitSemantic(commit.Message) {
			passMark = "x"
			terminalMessage = commit.Message
			failedCommitCount++
		}

		log.Logger().Infof("---  Commit | %s --- %s\n"+
			"%s",
			commit.Hash, passMark, terminalMessage)
	}

	if failedCommitCount > 0 {
		return fmt.Errorf("%d commit(s) did not follow https://conventionalcommits.org/, please rebase and merge", failedCommitCount)
	}
	log.Logger().Infof("\nAll commits follow Conventional Commits")
	return nil
}

// IsCommitSemantic checks whether the commit message follow the conventions set out in
// Conventional Commits
func IsCommitSemantic(commitMessage string) bool {
	commitMessage = strings.TrimSpace(strings.ToLower(commitMessage))

	idx := strings.Index(commitMessage, ":")
	if idx > 0 {
		commitTitle := commitMessage[0:idx]
		for _, semanticTitle := range ConventionalCommitTitles {
			if strings.HasPrefix(commitTitle, semanticTitle) {
				return true
			}
		}
	}

	// Check that the commit isn't a revert or merge
	if strings.Contains(commitMessage, "revert") || strings.Contains(commitMessage, "merge") {
		return true
	}
	return false
}

func (o *Options) git() gitclient.Interface {
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.CommandRunner)
	}
	return o.GitClient
}
