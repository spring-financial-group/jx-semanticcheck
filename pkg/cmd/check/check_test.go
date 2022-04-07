//go:build unit
// +build unit

package check

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"spring-financial-group/jx-semanticcheck/internal/domain/mocks"
	"testing"
)

func TestGetPreviousRevision_FirstRelease(t *testing.T) {
	mockGitClient := new(mocks.Interface)
	mockGitClient.On("Command", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("FirstCommitSHA", nil)
	t.Run("First Release", func(t *testing.T) {
		actualRevision, err := GetPreviousRevSha(mockGitClient, "")
		assert.NoError(t, err)
		assert.Equal(t, "FirstCommitSHA", actualRevision)
	})
}

func TestGetPreviousRevision_NthRelease(t *testing.T) {
	actualRevSHA := "SHA2"
	mockGitClient := new(mocks.Interface)
	mockGitClient.On("Command", mock.Anything, "for-each-ref", "--sort=-creatordate", "--format=%(objectname)%00%(refname:short)", "--count=2", "refs/tags").Return(fmt.Sprintf("Tag1\x00SHA1\nTag2\x00%s", actualRevSHA), nil)
	mockGitClient.On("Command", mock.Anything, "rev-list", "-n", "1", "Tag2").Return(actualRevSHA, nil)
	t.Run("Nth Release", func(t *testing.T) {
		actualRevision, err := GetPreviousRevSha(mockGitClient, "")
		assert.NoError(t, err)
		assert.Equal(t, "SHA2", actualRevision)
	})
}

func TestIsCommitSemantic(t *testing.T) {
	testCases := []struct {
		name           string
		commitMessage  string
		expectedResult bool
	}{
		{
			name:           "revert",
			commitMessage:  "Revert: these really bad commits",
			expectedResult: true,
		},
		{
			name:           "chore",
			commitMessage:  "chore: update deps",
			expectedResult: true,
		},
		{
			name:           "feat",
			commitMessage:  "feat: add great feature",
			expectedResult: true,
		},
		{
			name:           "refactor",
			commitMessage:  "Refactor: improve code base",
			expectedResult: true,
		},
		{
			name:           "docs",
			commitMessage:  "docs: improve docs",
			expectedResult: true,
		},
		{
			name:           "test",
			commitMessage:  "test: make all the tests",
			expectedResult: true,
		},
		{
			name:           "style",
			commitMessage:  "Style: improve the style",
			expectedResult: true,
		},
		{
			name:           "not semantic",
			commitMessage:  "improve this code",
			expectedResult: false,
		},
		{
			name:           "without colon",
			commitMessage:  "fix this code",
			expectedResult: false,
		},
		{
			name:           "with feature",
			commitMessage:  "docs(README): improve readability",
			expectedResult: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualResult := IsCommitSemantic(tt.commitMessage)
			assert.Equal(t, tt.expectedResult, actualResult)
		})
	}
}
