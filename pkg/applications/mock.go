package applications

import (
	"context"
	"strings"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var _ ApplicationGetter = (*MockGetter)(nil)

// NewMock returns a simple secret getter.
func NewMock() *MockGetter {
	return &MockGetter{}
}

// MockGetter implements the ApplicationGetter interface.
type MockGetter struct {
	list []appv1.Application
}

// Application implements the ApplicationGetter interface.
func (k *MockGetter) GetApplication(ctx context.Context, authToken string, id types.NamespacedName) (*appv1.Application, error) {
	// tokent, ok := k.secrets[mockKey(authToken, secretID, key)]
	// if !ok {
	// 	return "", fmt.Errorf("mock not found")
	// }

	// TODO: Fix this.
	return nil, nil
}

// TODO: add auth-token checking
func (k *MockGetter) ListApplications(ctx context.Context, _ string, namespace string) ([]appv1.Application, error) {
	r := []appv1.Application{}
	for _, v := range k.list {
		if v.ObjectMeta.Namespace == namespace {
			r = append(r, v)
		}
	}
	return r, nil
}

// AddListResponse is a mock method that adds to the list of apps to be
// returned.
func (k *MockGetter) AddListResponse(a appv1.Application) {
	k.list = append(k.list, a)
}

func mockKey(token string, n types.NamespacedName, key string) string {
	return strings.Join([]string{token, n.Name, n.Namespace, key}, ":")
}
