package applications

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

// ApplicationGetter takes a namespaced name and finds an ArgoCD application with that name, or
// returns an error.
type ApplicationGetter interface {
	GetApplication(ctx context.Context, authToken string, id types.NamespacedName) (*appv1.Application, error)
	ListApplications(ctx context.Context, authToken string, namespace string) ([]appv1.Application, error)
}
