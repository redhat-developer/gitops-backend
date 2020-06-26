package secrets

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubeSecretGetter is an implementation of SecretGetter.
type KubeSecretGetter struct {
	configFactory RESTConfigFactory
	clientFactory func(*rest.Config) (kubernetes.Interface, error)
}

// NewFromConfig creates a secret getter from a rest.Config.
func NewFromConfig(cfg *rest.Config, insecure bool) *KubeSecretGetter {
	return New(NewRESTConfigFactory(cfg, insecure))
}

// New creates and returns a KubeSecretGetter that looks up secrets in k8s.
func New(c RESTConfigFactory) *KubeSecretGetter {
	return &KubeSecretGetter{
		configFactory: c,
		clientFactory: func(c *rest.Config) (kubernetes.Interface, error) {
			return kubernetes.NewForConfig(c)
		},
	}
}

// SecretToken looks for a namespaced secret, and returns the 'token' key from
// it, or an error if not found.
func (k KubeSecretGetter) SecretToken(ctx context.Context, authToken string, id types.NamespacedName) (string, error) {
	cfg, err := k.configFactory.Create(authToken)
	if err != nil {
		return "", fmt.Errorf("failed to create a REST config: %w", err)
	}
	coreClient, err := k.clientFactory(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create a client from the config: %w", err)
	}
	secret, err := coreClient.CoreV1().Secrets(id.Namespace).Get(ctx, id.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting secret %s/%s: %w", id.Namespace, id.Name, err)
	}
	token, ok := secret.Data["token"]
	if !ok {
		return "", fmt.Errorf("secret invalid, no 'token' key in %s/%s", id.Namespace, id.Name)
	}
	return string(token), nil
}
