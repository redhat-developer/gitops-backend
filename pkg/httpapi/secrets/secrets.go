package secrets

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// KubeSecretGetter is an implementation of SecretGetter.
type KubeSecretGetter struct {
	coreClient kubernetes.Interface
}

// New creates and returns a KubeSecretGetter that looks up secrets in k8s.
func New(c kubernetes.Interface) *KubeSecretGetter {
	return &KubeSecretGetter{
		coreClient: c,
	}
}

// SecretToken looks for a namespaced secret, and returns the 'token' key from
// it, or an error if not found.
func (k KubeSecretGetter) SecretToken(ctx context.Context, id types.NamespacedName) (string, error) {
	secret, err := k.coreClient.CoreV1().Secrets(id.Namespace).Get(ctx, id.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting secret %s/%s: %w", id.Namespace, id.Name, err)
	}
	token, ok := secret.Data["token"]
	if !ok {
		return "", fmt.Errorf("secret invalid, no 'token' key in %s/%s", id.Namespace, id.Name)
	}
	return string(token), nil
}
