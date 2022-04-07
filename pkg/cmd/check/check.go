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

	Namespace           string
	LatestCommitSha     string
	PreviousRevisionSha string
	RepoDir             string
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
	cmd.Flags().StringVarP(&o.PreviousRevisionSha, "previous-rev", "p", "", "the previous tag revision")
	cmd.Flags().StringVarP(&o.LatestCommitSha, "latest-sha", "", "", "the current tag revision")
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

	dir := o.RepoDir
	if dir == "" {
		dir = o.ScmFactory.Dir
	}

	if o.PreviousRevisionSha == "" {
		o.PreviousRevisionSha, err = GetPreviousRevSha(o.git(), dir)
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

	commits, err := chgit.FetchCommits(dir, o.PreviousRevisionSha, o.LatestCommitSha)
	if err != nil {
		return err
	}

	commitSlice := *commits
	if strings.HasPrefix(commitSlice[0].Message, "release ") {
		// remove the release commit from the log
		commitSlice = commitSlice[1:]
	}

	var semanticCounter int
	for _, commit := range commitSlice {
		if !IsCommitSemantic(commit.Message) {
			log.Logger().Infof("commit %s does not use Conventional Commits:\n%s\n", commit.Hash, commit.Message)
			semanticCounter++
		}
	}
	if semanticCounter > 0 {
		return fmt.Errorf("%d commit(s) did not follow Conventional Commits, please rebase and merge", semanticCounter)
	}
	log.Logger().Infof("all commits follow Conventional Commits")
	return nil
}

func GetPreviousRevSha(g gitclient.Interface, dir string) (string, error) {
	previousRevSha, _, err := gits.GetCommitPointedToByPreviousTag(g, dir)
	if err != nil {
		return "", err
	}
	if previousRevSha == "" {
		// let's assume we are the first release
		previousRevSha, err = gits.GetFirstCommitSha(g, dir)
		if err != nil {
			return "", errors.Wrap(err, "failed to find first commit after we found no previous releases")
		}
		if previousRevSha == "" {
			return "", errors.Errorf("no previous commit version found so change diff unavailable")
		}
	}
	return previousRevSha, nil
}

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
	return false
}

func (o *Options) git() gitclient.Interface {
	if o.GitClient == nil {
		o.GitClient = cli.NewCLIClient("", cmdrunner.QuietCommandRunner)
	}
	return o.GitClient
}
