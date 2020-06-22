package secrets

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

var _ SecretGetter = (*KubeSecretGetter)(nil)

func TestSecret(t *testing.T) {
	id := types.NamespacedName{Name: "test-secret", Namespace: "test-ns"}
	fakeClient := fake.NewSimpleClientset(createSecret(id, "secret-token"))
	g := New(fakeClient)

	secret, err := g.SecretToken(context.TODO(), id)
	if err != nil {
		t.Fatal(err)
	}

	if secret != "secret-token" {
		t.Fatalf("got %s, want secret-token", secret)
	}
}

func TestSecretWithMissingSecret(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	g := New(fakeClient)

	id := types.NamespacedName{Name: "test-secret", Namespace: "test-ns"}
	_, err := g.SecretToken(context.TODO(), id)
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
