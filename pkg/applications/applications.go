package applications

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	appclientset "github.com/argoproj/argo-cd/pkg/client/clientset/versioned"
	"github.com/redhat-developer/gitops-backend/pkg/clients"
)

// KubeApplicationGetter is an implementation of ApplicationGetter.
type KubeApplicationGetter struct {
	configFactory clients.RESTConfigFactory
	clientFactory func(*rest.Config) (appclientset.Interface, error)
}

// NewFromConfig creates an application getter from a rest.Config.
func NewFromConfig(cfg *rest.Config, insecure bool) *KubeApplicationGetter {
	return New(clients.NewRESTConfigFactory(cfg, insecure))
}

// New creates and returns a KubeApplicationGetter that looks up secrets in k8s.
func New(c clients.RESTConfigFactory) *KubeApplicationGetter {
	return &KubeApplicationGetter{
		configFactory: c,
		clientFactory: func(c *rest.Config) (appclientset.Interface, error) {
			return appclientset.NewForConfig(c)
		},
	}
}

// GetApplication looks for a namespaced application and returns it or an error if
// not found.
func (k KubeApplicationGetter) GetApplication(ctx context.Context, authToken string, id types.NamespacedName) (*appv1.Application, error) {
	appClient, err := k.client(authToken)
	if err != nil {
		return nil, err
	}
	app, err := appClient.ArgoprojV1alpha1().Applications(id.Namespace).Get(ctx, id.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting application %s: %w", id, err)
	}
	return app, nil
}

func (k KubeApplicationGetter) ListApplications(ctx context.Context, authToken string, namespace string) ([]appv1.Application, error) {
	appClient, err := k.client(authToken)
	if err != nil {
		return nil, err
	}
	list, err := appClient.ArgoprojV1alpha1().Applications(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing applications %s: %w", namespace, err)
	}
	return list.Items, nil
}

func (k KubeApplicationGetter) client(authToken string) (appclientset.Interface, error) {
	cfg, err := k.configFactory.Create(authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create a REST config: %w", err)
	}
	appClient, err := k.clientFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create a client from the config: %w", err)
	}
	return appClient, nil
}
