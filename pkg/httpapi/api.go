package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi/secrets"
	"github.com/julienschmidt/httprouter"
	"sigs.k8s.io/yaml"
)

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
	scmClient git.SCM
	secrets   secrets.SecretGetter
}

// GePipelines fetches and returns the pipeline body.
func (a *APIRouter) GetPipelines(w http.ResponseWriter, r *http.Request) {
	urlToFetch := r.URL.Query().Get("url")
	if urlToFetch == "" {
		http.Error(w, "missing parameter 'url'", http.StatusBadRequest)
		return
	}
	repo, err := parseURL(urlToFetch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := a.scmClient.FileContents(r.Context(), repo, "pipelines.yaml", "master")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pipelines := &config{}
	err = yaml.Unmarshal(body, &pipelines)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal pipelines.yaml: %s", err.Error()), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(pipelinesToAppsResponse(pipelines))
}

// NewRouter creates and returns a new APIRouter.
func NewRouter(c git.SCM, s secrets.SecretGetter) *APIRouter {
	api := &APIRouter{Router: httprouter.New(), scmClient: c, secrets: s}
	api.HandlerFunc(http.MethodGet, "/pipelines", api.GetPipelines)
	return api
}

func parseURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse %#v: %w", s, err)
	}
	return strings.TrimLeft(strings.Trim(parsed.Path, ".git"), "/"), nil
}

// TODO: this should really import the config from the upstream and use it to
// unmarshal.
func pipelinesToAppsResponse(cfg *config) *appsResponse {
	appSet := map[string]bool{}
	for _, env := range cfg.Environments {
		for _, app := range env.Apps {
			appSet[app.Name] = true
		}
	}

	apps := []appResponse{}
	for k := range appSet {
		apps = append(apps, appResponse{Name: k, RepoURL: cfg.GitOpsURL})
	}
	return &appsResponse{Apps: apps}
}

type appResponse struct {
	Name    string `json:"name,omitempty"`
	RepoURL string `json:"repo_url,omitempty"`
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
