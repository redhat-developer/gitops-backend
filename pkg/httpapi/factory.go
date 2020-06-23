package http

import (
	"fmt"

	scmfactory "github.com/jenkins-x/go-scm/scm/factory"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/metrics"
)

// SCMClientFactory is an implementation of the ClientFactory interface that can
// create clients based on go-scm.
type SCMClientFactory struct {
	drivers DriverIdentifier
	metrics metrics.Interface
}

// NewClientFactory creates and returns an SCMClientFactory.
func NewClientFactory(d DriverIdentifier, m metrics.Interface) *SCMClientFactory {
	return &SCMClientFactory{drivers: d, metrics: m}
}

func (s *SCMClientFactory) Create(url, token string) (git.SCM, error) {
	driver, err := s.drivers.Identify(url)
	if err != nil {
		return nil, err
	}
	scmClient, err := scmfactory.NewClient(driver, "", token)
	if err != nil {
		return nil, fmt.Errorf("failed to create a git driver: %s", err)
	}
	return git.New(scmClient, s.metrics), nil
}
