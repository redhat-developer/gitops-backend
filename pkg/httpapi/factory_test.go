package http

import (
	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi/secrets"
)

type SCMClientFactory struct {
	secrets secrets.SecretGetter
	drivers DriverIdentifier
}

func (s *SCMClientFactory) Create(url, token string) (git.SCM, error) {
	return nil, nil
}

var _ ClientFactory = (*SCMClientFactory)(nil)
