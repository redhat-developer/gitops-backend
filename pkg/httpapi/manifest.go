package httpapi

type appResponse struct {
	Name         string   `json:"name,omitempty"`
	RepoURL      string   `json:"repo_url,omitempty"`
	Environments []string `json:"environments,omitempty"`
}

type appsResponse struct {
	Apps []appResponse `json:"applications"`
}

type config struct {
	GitOpsURL    string         `json:"gitops_url"`
	Environments []*environment `json:"environments,omitempty"`
}

type environment struct {
	Name    string         `json:"name,omitempty"`
	Cluster string         `json:"cluster,omitempty"`
	Apps    []*application `json:"apps,omitempty"`
}

type application struct {
	Name string `json:"name,omitempty"`
}

func (c *config) findEnvironment(n string) *environment {
	for _, e := range c.Environments {
		if e.Name == n {
			return e
		}
	}
	return nil
}
