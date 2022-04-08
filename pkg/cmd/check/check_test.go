package check

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
