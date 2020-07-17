package httpapi

import (
	"fmt"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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
		"resources":   res,
	}
	return appEnv, nil
}

func pathForApplication(appName, envName string) string {
	return path.Join("environments", envName, "apps", appName)
}
