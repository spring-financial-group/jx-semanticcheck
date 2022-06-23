package helpers

import (
	"fmt"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/pkg/errors"
	"strings"
)

type Commit struct {
	SHA     string
	Message string
	Date    string
}

// GetNewCommits returns a list of Commit that have yet to be applied upstream
// from the current branch
func GetNewCommits(gitter gitclient.Interface, dir string) ([]*Commit, error) {
	defaultBranch, err := getDefaultBranchName(gitter, dir)
	if err != nil {
		return nil, err
	}

	// Gets a list of the commits on the current branch and whether they are new
	out, err := gitter.Command(dir, "cherry", fmt.Sprintf("origin/%s", defaultBranch))
	if err != nil {
		return nil, errors.Wrapf(err, "running git")
	}
	split := strings.Split(out, "\n")

	var newCommits []string
	for _, hash := range split {
		// Filter newCommits for commits that are present upstream
		if strings.Contains(hash, "+") {
			hash = strings.ReplaceAll(hash, "+ ", "")
			newCommits = append(newCommits, hash)
		}
	}

	return GetCommits(gitter, dir, newCommits)
}

func GetCommits(gitter gitclient.Interface, dir string, SHAs []string) ([]*Commit, error) {
	var commits []*Commit
	for _, sha := range SHAs {
		commit, err := GetCommit(gitter, dir, sha)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// GetCommit uses "git show" to get information about a specific commit
func GetCommit(gitter gitclient.Interface, dir string, SHA string) (*Commit, error) {
	out, err := gitter.Command(dir, "show", "--quiet", fmt.Sprintf("%s", SHA),
		"--format=%H%n%s%n%ai")
	if err != nil {
		return nil, errors.Wrapf(err, "running git")
	}
	split := strings.Split(out, "\n")
	return &Commit{
		SHA:     split[0],
		Message: split[1],
		Date:    split[2],
	}, nil
}

func getDefaultBranchName(gitter gitclient.Interface, dir string) (string, error) {
	out, err := gitter.Command(dir, "ls-remote", "--symref", "origin",
		"HEAD")
	if err != nil {
		return "", errors.Wrapf(err, "running git")
	}
	branchName := "master"
	if strings.Contains(out, "main") {
		branchName = "main"
	}
	return branchName, nil
}
