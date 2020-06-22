package git

import (
	"context"
)

// SCM is a wrapper around go-scm's Client implementation.
type SCM interface {
	// FileContents returns the contents of a file within a repo.
	FileContents(ctx context.Context, repo, path, ref string) ([]byte, error)
}
