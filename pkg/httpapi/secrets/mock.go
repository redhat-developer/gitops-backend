package secrets

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

var _ SecretGetter = (*MockSecret)(nil)

// NewMock returns a simple secret getter.
func NewMock() MockSecret {
	return MockSecret{}
}

// MockSecret implements the SecretGetter interface.
type MockSecret struct {
	secrets map[string]string
}

// Secret implements the SecretGetter interface.
func (k MockSecret) SecretToken(ctx context.Context, authToken string, secretID types.NamespacedName) (string, error) {
	token, ok := k.secrets[key(authToken, secretID)]
	if !ok {
		return "", fmt.Errorf("mock not found")
	}
	return token, nil
}

// AddStubResponse is a mock method that sets up a token to be returned.
func (k MockSecret) AddStubResponse(authToken string, secretID types.NamespacedName, token string) {
	k.secrets[key(authToken, secretID)] = token
}

func key(token string, n types.NamespacedName) string {
	return fmt.Sprintf("%s:%s:%s", token, n.Name, n.Namespace)
}
