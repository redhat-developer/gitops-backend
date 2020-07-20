package argocd

import (
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

// ArgoCDClient implementations can fetch a list of resources from ArgoCD for an
// application.
type ArgoCDClient interface {
	ApplicationResources(name string) ([]*resource.Resource, error)
}
