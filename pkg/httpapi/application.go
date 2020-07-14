package httpapi

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/gitfs"
)

func applicationEnvironment(c *config, appName, envName string) (map[string]string, error) {
	appEnv := map[string]string{
		"environment": envName,
	}
	gfs, err := gitfs.NewInMemoryFromOptions(&git.CloneOptions{
		URL: c.GitOpsURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a GitFS frm %s: %w", c.GitOpsURL, err)
	}

	return appEnv, nil
}
