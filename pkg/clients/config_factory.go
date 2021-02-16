package clients

import (
	"k8s.io/client-go/rest"
)

// K8sRESTConfigFactory is an implementation of the RESTConfigFactory interface
// that doles out configs based on a base one.
type K8sRESTConfigFactory struct {
	insecure bool
	cfg      *rest.Config
}

// Create implements the RESTConfigFactory interface.
func (r *K8sRESTConfigFactory) Create(token string) (*rest.Config, error) {
	shallowCopy := *r.cfg
	shallowCopy.BearerToken = token
	if r.insecure {
		shallowCopy.TLSClientConfig = rest.TLSClientConfig{
			Insecure: r.insecure,
		}
	}
	return &shallowCopy, nil
}

// NewRestConfigFactory creates and returns a RESTConfigFactory with a known
// config.
func NewRESTConfigFactory(cfg *rest.Config, insecure bool) *K8sRESTConfigFactory {
	return &K8sRESTConfigFactory{cfg: cfg, insecure: insecure}
}
