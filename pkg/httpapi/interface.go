package httpapi

import (
	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
)

// GitClientFactory is an interface for creating git.SCM clients based on the URL
// to be fetched.
type GitClientFactory interface {
	// Create creates a new client, using the provided token for authentication.
	Create(url, token string) (git.SCM, error)
}

// DriverIdentifer parses a URL and attempts to determine which go-scm driver to
// use to talk to the server.
type DriverIdentifier interface {
	Identify(url string) (string, error)
}
