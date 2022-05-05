package httpapi

type appResponse struct {
	Name         string   `json:"name,omitempty"`
	RepoURL      string   `json:"repo_url,omitempty"`
	Environments []string `json:"environments,omitempty"`
	SyncStatus   []string `json:"sync_status,omitempty"`
	LastDeployed []string `json:"last_deployed,omitempty"`
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
	Name     string    `json:"name,omitempty"`
	Services []service `json:"services,omitempty"`
}

type envHealthResource struct {
	Name   string `json:"name,omitempty"`
	Health string `json:"health,omitempty"`
	Status string `json:"status,omitempty"`
}

type envHistory struct {
	Author      string `json:"author,omitempty"`
	Message     string `json:"message,omitempty"`
	Revision    string `json:"revision,omitempty"`
	Environment string `json:"environment,omitempty"`
	RepoUrl     string `json:"repo_url,omitempty"`
	DeployedAt  string `json:"deployed_at,omitempty"`
}

func (e environment) findService(n string) *service {
	for _, a := range e.Apps {
		for _, s := range a.Services {
			if s.Name == n {
				return &s
			}
		}
	}
	return nil
}

type service struct {
	Name      string `json:"name"`
	SourceURL string `json:"source_url"`
}

func (c *config) findEnvironment(n string) *environment {
	for _, e := range c.Environments {
		if e.Name == n {
			return e
		}
	}
	return nil
}
