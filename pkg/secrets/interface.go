package secrets

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

// SecretGetter takes a namespaced name and finds a secret with that name, or
// returns an error.
type SecretGetter interface {
	SecretToken(ctx context.Context, authToken string, id types.NamespacedName, key string) (string, error)
}

// RESTConfigFactory creates and returns new Kubernetes client configurations
// for accessing the API.
type RESTConfigFactory interface {
	Create(token string) (*rest.Config, error)
}
