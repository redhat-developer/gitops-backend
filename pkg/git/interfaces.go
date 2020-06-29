package git

import (
	"context"
)

// ClientFactory is an interface for creating SCM clients based on the URL
// to be fetched.
type ClientFactory interface {
	// Create creates a new client, using the provided token for authentication.
	Create(url, token string) (SCM, error)
}

// DriverIdentifer parses a URL and attempts to determine which go-scm driver to
// use to talk to the server.
type DriverIdentifier interface {
	Identify(url string) (string, error)
}

// SCM is a wrapper around go-scm's Client implementation.
type SCM interface {
	// FileContents returns the contents of a file within a repo.
	FileContents(ctx context.Context, repo, path, ref string) ([]byte, error)
}
