package test

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// MakeCloneOptions determines if we are running in Travis,
// and ensures that it's using the correct branch.
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
