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

func parseServicesFromResources(env *environment, res []*resource.Resource) []responseService {
	serviceImages := map[string][]string{}
	serviceResources := map[string][]*resource.Resource{}
	for _, v := range res {
		name := serviceFromLabels(v.Labels)

		images, ok := serviceImages[name]
		if !ok {
			images = []string{}
		}
		images = append(images, v.Images...)
		serviceImages[name] = images

		resources, ok := serviceResources[name]
		if !ok {
			resources = []*resource.Resource{}
		}
		resources = append(resources, v)
		serviceResources[name] = resources
	}

	services := []responseService{}
	for k, v := range serviceImages {
		svc := env.findService(k)
		svcRepo := ""
		if svc != nil {
			svcRepo = svc.SourceURL
		}
		services = append(services, responseService{
			Name:      k,
			Images:    v,
			Source:    source{URL: svcRepo},
			Resources: serviceResources[k]})
	}

	return services
}

const nameLabel = "app.kubernetes.io/name"

func serviceFromLabels(l map[string]string) string {
	return l[nameLabel]
}

type responseService struct {
	Name      string               `json:"name"`
	Source    source               `json:"source,omitempty"`
	Images    []string             `json:"images,omitempty"`
	Badge     string               `json:"badge"`
	Resources []*resource.Resource `json:"resources,omitempty"`
}

type source struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}
