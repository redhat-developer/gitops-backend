package argocd

import (
	"crypto/tls"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client"
	// appsvc "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client/application_service"
)

type ArgoCDClient interface {
}

type ArgoCD struct {
	apiClient *apiclient.Argocd
}

func (a ArgoCD) Application(name string) string {
	params := appsvc.NewGetMixin8Params()
	params = params.WithName(viper.GetString(applicationFlag))

	// 	res, err := c.ApplicationService.GetMixin8(params)
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
	return ArgoCD{apiClient: apiclient.New(r, strfmt.Default)}
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
