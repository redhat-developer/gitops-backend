package secrets

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

var _ SecretGetter = (*KubeSecretGetter)(nil)

var testID = types.NamespacedName{Name: "test-secret", Namespace: "test-ns"}

func TestSecret(t *testing.T) {
	g := New(&stubConfigFactory{})
	g.clientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		if c.BearerToken == "auth token" {
			return fake.NewSimpleClientset(createSecret(testID, "secret-token")), nil
		}
		return nil, errors.New("failed")
	}

	secret, err := g.SecretToken(context.TODO(), "auth token", testID, "token")
	if err != nil {
		t.Fatal(err)
	}

	if secret != "secret-token" {
		t.Fatalf("got %s, want secret-token", secret)
	}
}

func TestSecretWithMissingSecret(t *testing.T) {
	g := New(&stubConfigFactory{})
	g.clientFactory = func(c *rest.Config) (kubernetes.Interface, error) {
		if c.BearerToken == "auth token" {
			return fake.NewSimpleClientset(), nil
		}
		return nil, errors.New("failed")
	}

	_, err := g.SecretToken(context.TODO(), "auth token", testID, "token")
	if err.Error() != `error getting secret test-ns/test-secret: secrets "test-secret" not found` {
		t.Fatal(err)
	}
}

func createSecret(id types.NamespacedName, token string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      id.Name,
			Namespace: id.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(token),
		},
	}
}

type stubConfigFactory struct {
}

func (s *stubConfigFactory) Create(token string) (*rest.Config, error) {
	return &rest.Config{BearerToken: token}, nil
}
