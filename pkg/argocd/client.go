package argocd

import (
	"crypto/tls"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client"
	appsvc "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client/application_service"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

type argoAPIClient interface {
}

type ArgoCDClient interface {
	ApplicationResources(name string) ([]*resource.Resource, error)
}

type ArgoCD struct {
	apiClient *apiclient.Argocd
}

func (a ArgoCD) ApplicationResources(name string) ([]*resource.Resource, error) {
	params := appsvc.NewGetMixin8Params().WithName(name)
	res, err := a.apiClient.ApplicationService.GetMixin8(params)
	if err != nil {
		return nil, err
	}
	resources := []*resource.Resource{}

	for _, v := range res.Payload.Status.Resources {
		resources = append(resources, &resource.Resource{
			Group:     v.Group,
			Version:   v.Version,
			Kind:      v.Kind,
			Name:      v.Name,
			Namespace: v.Namespace,
		})
	}
	return resources, nil
}

// New creates a new ArgoCD API Client.
//
// endpoint should be of the form hostname:port e.g. argocd.svc:8080
// the token should be a generated API token created using
// https://argoproj.github.io/argo-cd/developer-guide/api-docs/
// insecure should be true, if you don't want to validate the TLS connection
// not recommended - this is very insecure
func New(endpoint, token string, insecure bool) *ArgoCD {
	r := httptransport.NewWithClient(endpoint, apiclient.DefaultBasePath, apiclient.DefaultSchemes, tlsClient(insecure))
	bearerTokenAuth := httptransport.BearerToken(token)
	r.DefaultAuthentication = bearerTokenAuth
	return NewFromClient(apiclient.New(r, strfmt.Default))
}

// NewFromClient creates a new ArgoCD client from an existing api client.
func NewFromClient(a *apiclient.Argocd) *ArgoCD {
	return &ArgoCD{apiClient: a}
}

func tlsClient(insecure bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
}
