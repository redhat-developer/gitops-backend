package test

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// MakeCloneOptions determines if we are running in GitHub,
// and ensures that it's using the correct branch.
//
// This is because the codebase that the tests run in is not a clone, so we have
// to clone the upstream.
//
// For local development, you're working on the branch, and we can use that
// branch correctly.
func MakeCloneOptions() *git.CloneOptions {
	o := &git.CloneOptions{
		URL:   "../..",
		Depth: 1,
	}
	if b := os.Getenv("GITHUB_BASE_REF"); b != "" {
		o.ReferenceName = plumbing.NewBranchReferenceName(b)
		o.URL = "https://github.com/redhat-developer/gitops-backend.git"
	}
	return o
}
