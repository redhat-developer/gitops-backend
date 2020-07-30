package httpapi

import (
	"fmt"
	"net/url"
	"path"
	"sort"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rhd-gitops-example/gitops-backend/pkg/resource"
)

// TODO: if the environment doesn't exist, this should return a not found error.
func (a *APIRouter) environmentApplication(authToken string, c *config, envName, appName string) (map[string]interface{}, error) {
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
	services, err := parseServicesFromResources(env, res)
	if err != nil {
		return nil, err
	}
	appEnv := map[string]interface{}{
		"environment": envName,
		"cluster":     c.findEnvironment(envName).Cluster,
		"services":    services,
	}
	return appEnv, nil
}

func pathForApplication(appName, envName string) string {
	return path.Join("environments", envName, "apps", appName)
}

func parseServicesFromResources(env *environment, res []*resource.Resource) ([]responseService, error) {
	serviceImages := map[string]map[string]bool{}
	serviceResources := map[string][]*resource.Resource{}
	for _, v := range res {
		name := serviceFromLabels(v.Labels)
		images, ok := serviceImages[name]
		if !ok {
			images = map[string]bool{}
		}
		for _, n := range v.Images {
			images[n] = true
		}
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
		rs := responseService{
			Name:      k,
			Images:    keys(v),
			Resources: serviceResources[k],
		}
		if svcRepo != "" {
			domain, err := hostFromURL(svcRepo)
			if err != nil {
				return nil, err
			}
			rs.Source = source{URL: svcRepo, Type: domain}
		}
		services = append(services, rs)
	}
	return services, nil
}

const nameLabel = "app.kubernetes.io/name"

func serviceFromLabels(l map[string]string) string {
	return l[nameLabel]
}

type responseService struct {
	Name      string               `json:"name"`
	Source    source               `json:"source,omitempty"`
	Images    []string             `json:"images,omitempty"`
	Resources []*resource.Resource `json:"resources,omitempty"`
}

type source struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

func hostFromURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("failed to parse Git repo URL %q: %w", u, err)
	}
	return parsed.Host, nil
}

func keys(v map[string]bool) []string {
	keys := []string{}
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
