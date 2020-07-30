package git

import (
	"fmt"
	"net/url"

	scmfactory "github.com/jenkins-x/go-scm/scm/factory"
	"github.com/rhd-gitops-example/gitops-backend/pkg/metrics"
)

// SCMClientFactory is an implementation of the GitClientFactory interface that can
// create clients based on go-scm.
type SCMClientFactory struct {
	metrics metrics.Interface
}

// NewClientFactory creates and returns an SCMClientFactory.
func NewClientFactory(m metrics.Interface) *SCMClientFactory {
	return &SCMClientFactory{metrics: m}
}

func (s *SCMClientFactory) Create(url, token string) (SCM, error) {
	host, err := hostFromURL(url)
	if err != nil {
		return nil, err
	}
	driver, err := scmfactory.DefaultIdentifier.Identify(host)
	if err != nil {
		return nil, err
	}
	scmClient, err := scmfactory.NewClient(driver, "", token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a git driver: %w", err)
	}
	return New(scmClient, s.metrics), nil
}

func hostFromURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", s, err)
	}
	return parsed.Host, nil
}
