package httpapi

import (
	"fmt"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

// TODO: if the environment doesn't exist, this should return a not found error.
func (a *APIRouter) applicationEnvironment(authToken string, c *config, appName, envName string) (map[string]interface{}, error) {
	if c.GitOpsURL == "" {
		return nil, nil
	}
	env := c.findEnvironment(envName)
	if env == nil {
		return nil, fmt.Errorf("failed to find environment %#v", envName)
	}
	co := &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "gitops",
			Password: authToken,
		},
		URL: c.GitOpsURL,
	}
	res, err := a.resourceParser(pathForApplication(appName, envName), co)
	if err != nil {
		return nil, err
	}
	appEnv := map[string]interface{}{
		"environment": envName,
		"cluster":     c.findEnvironment(envName).Cluster,
		"services":    parseServicesFromResources(env, res),
	}
	return appEnv, nil
}

func pathForApplication(appName, envName string) string {
	return path.Join("environments", envName, "apps", appName)
}

func parseServicesFromResources(env *environment, res []*resource.Resource) []service {
	return nil
}

type service struct {
	Name      string              `json:"name"`
	Source    source              `json:"source,omitempty"`
	Images    []string            `json:"images,omitempty"`
	Badge     string              `json:"badge"`
	Resources []resource.Resource `json:"resources,omitempty"`
}

type source struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// {
//   "cluster": "http://dev-stage-cluster.com",
//   "environment": "dev",
//   "namespace": "dev-ns",
//   "services": [
//     {
//       "name": "taxi-frontend",
//       "source": {
//         "url": "https://github.com/taxi-frontend",
//         "type": "git"
//       },
//       "images": [
//         "nodejs:latest",
//         "service:9326bb256681145ce73408e331513e87"
//       ],
//       "badge": "nodejs",
//       "resources": [
//         {
//           "group": "apps",
//           "version": "v1",
//           "kind": "Deployment",
//           "name": "taxi-frontend-demo-svc",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "",
//           "version": "v1",
//           "kind": "Service",
//           "name": "taxi-frontend-demo-svc",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "route.openshift.io",
//           "version": "v1",
//           "kind": "Route",
//           "name": "taxi-frontend-demo-svc-route",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "rbac.authorization.k8s.io",
//           "version": "v1",
//           "kind": "RoleBinding",
//           "name": "dev-rolebinding",
//           "namespace": "dev-ns"
//         }
//       ]
//     },
//     {
//       "name": "taxi-backend",
//       "source": {
//         "url": "https://github.com/taxi-backend",
//         "type": "git"
//       },
//       "images": [
//         "http-api:v1",
//         "redis:latest",
//         "service:9326bb256681145ce73408e331513e87"
//       ],
//       "badge": "go",
//       "resources": [
//         {
//           "group": "apps",
//           "version": "v1",
//           "kind": "Deployment",
//           "name": "taxi-backend-demo-svc",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "",
//           "version": "v1",
//           "kind": "Service",
//           "name": "tax-backend-demo-svc",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "route.openshift.io",
//           "version": "v1",
//           "kind": "Route",
//           "name": "taxi-backend-demo-svc-route",
//           "namespace": "dev-ns"
//         },
//         {
//           "group": "rbac.authorization.k8s.io",
//           "version": "v1",
//           "kind": "RoleBinding",
//           "name": "dev-rolebinding",
//           "namespace": "dev-ns"
//         }
//       ]
//     }
//   ]
// }
