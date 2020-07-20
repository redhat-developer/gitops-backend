package parser

import (
	"github.com/go-git/go-git/v5"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

// ResourceParser implementations should fetch the source from the repoURL and
// parse the resources in the path into a set of git.Resource values.
type ResourceParser func(path string, opts *git.CloneOptions) ([]*resource.Resource, error)
