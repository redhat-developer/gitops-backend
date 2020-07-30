package git

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/rhd-gitops-example/gitops-backend/pkg/metrics"
)

// New creates and returns a new SCMClient.
func New(c *scm.Client, m metrics.Interface) *SCMClient {
	return &SCMClient{Client: c, m: m}
}

// SCMClient is a wrapper for the go-scm scm.Client with a simplified API.
type SCMClient struct {
	Client *scm.Client
	m      metrics.Interface
}

// FileContents reads the specific revision of a file from a repository.
//
// If an HTTP error is returned by the upstream service, an error with the
// response status code is returned.
func (c *SCMClient) FileContents(ctx context.Context, repo, path, ref string) ([]byte, error) {
	c.m.CountAPICall("file_contents")
	content, r, err := c.Client.Contents.Find(ctx, repo, path, ref)
	if r != nil && isErrorStatus(r.Status) {
		c.m.CountFailedAPICall("file_contents")
		return nil, SCMError{msg: fmt.Sprintf("failed to get file %s from repo %s ref %s", path, repo, ref), Status: r.Status}
	}
	if err != nil {
		c.m.CountFailedAPICall("file_contents")
		return nil, err
	}

	return content.Data, nil
}

func isErrorStatus(i int) bool {
	return i >= 400
}
