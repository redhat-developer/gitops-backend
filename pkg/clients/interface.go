package clients

import (
	"k8s.io/client-go/rest"
)

// RESTConfigFactory creates and returns new Kubernetes client configurations
// for accessing the API.
type RESTConfigFactory interface {
	Create(token string) (*rest.Config, error)
}
