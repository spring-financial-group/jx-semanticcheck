package domain

import "github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"

type Interface interface {
	gitclient.Interface
}
