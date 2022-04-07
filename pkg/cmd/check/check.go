package check

import (
	"fmt"
	chgit "github.com/antham/chyle/chyle/git"
	jxc "github.com/jenkins-x/jx-api/v4/pkg/client/clientset/versioned"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cmdrunner"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/helper"
	"github.com/jenkins-x/jx-helpers/v3/pkg/cobras/templates"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/jxclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"spring-financial-group/jx-semanticcheck/pkg/gits"
	"strings"
)

// Options contains the command line flags
type Options struct {
	options.BaseOptions

	ScmFactory    scmhelpers.Options
	GitClient     gitclient.Interface
	CommandRunner cmdrunner.CommandRunner
	JXClient      jxc.Interface

	Namespace        string
	CurrentRevision  string
	PreviousRevision string
}

var (
	cmdLong = templates.LongDesc(`
`)

	cmdExample = templates.Examples(`
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
	cmd.Flags().StringVarP(&o.PreviousRevision, "previous-rev", "p", "", "the previous tag revision")
	cmd.Flags().StringVarP(&o.CurrentRevision, "rev", "", "", "the current tag revision")

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

	o.JXClient, o.Namespace, err = jxclient.LazyCreateJXClientAndNamespace(o.JXClient, o.Namespace)
	if err != nil {
		return errors.Wrapf(err, "failed to create jx client")
	}
	return nil
}

func (o *Options) Run() error {
	err := o.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate")
	}

	dir := o.ScmFactory.Dir

	commits, err := GetCommitsSinceLastRevision(dir, o.git(), o.PreviousRevision, o.CurrentRevision)
	if err != nil {
		return err
	}

	commitSlice := *commits
	if strings.HasPrefix(commitSlice[0].Message, "release ") {
		// remove the release commit from the log
		tmp := commitSlice[1:]
		commits = &tmp
	}

	var semanticCounter int
	for _, commit := range commitSlice {
		if !IsCommitSemantic(commit.Message) {
			log.Logger().Errorf("commit %s does not use Conventional Commits:\n%s\n", commit.Hash, commit.Message)
			semanticCounter++
		}
	}
	if semanticCounter > 0 {
		return fmt.Errorf("%d commits did not follow Conventional Commits, please rebase and merge", semanticCounter+1)
	}
	return nil
}

func IsCommitSemantic(commitMessage string) bool {
	commitMessage = strings.TrimSpace(strings.ToLower(commitMessage))
	for _, title := range ConventionalCommitTitles {
		if strings.HasPrefix(commitMessage, title) {
			return true
		}
	}
	return false
}

func GetCommitsSinceLastRevision(dir string, g gitclient.Interface, previousRevision string, currentRevision string) (*[]object.Commit, error) {
	var err error
	if previousRevision == "" {
		previousRevision, _, err = gits.GetCommitPointedToByPreviousTag(g, dir)
		if err != nil {
			return nil, err
		}
		if previousRevision == "" {
			// lets assume we are the first release
			previousRevision, err = gits.GetFirstCommitSha(g, dir)
			if err != nil {
				return nil, errors.Wrap(err, "failed to find first commit after we found no previous releases")
			}
			if previousRevision == "" {
				return nil, errors.Errorf("no previous commit version found so change diff unavailable")
			}
		}
	}

	if currentRevision == "" {
		currentRevision, _, err = gits.GetCommitPointedToByLatestTag(g, dir)
		if err != nil {
			return nil, err
		}
	}

	gitDir, gitConfDir, err := gitclient.FindGitConfigDir(dir)
	if err != nil {
		return nil, err
	}
	if gitDir == "" || gitConfDir == "" {
		return nil, fmt.Errorf("no git directory could be found from dir %s", dir)
	}

	commits, err := chgit.FetchCommits(gitDir, previousRevision, currentRevision)
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (o *Options) git() gitclient.Interface {
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", o.CommandRunner)
	}
	return o.GitClient
}
