package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bigkevmcd/gitops-backend/pkg/git"
	"github.com/julienschmidt/httprouter"
	"sigs.k8s.io/yaml"
)

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
	scmClient git.SCM
}

// GePipelines fetches and returns the pipeline body.
func (a *APIRouter) GetPipelines(w http.ResponseWriter, r *http.Request) {
	urlToFetch := r.URL.Query().Get("url")
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
	response := map[string]interface{}{}
	err = yaml.Unmarshal(body, &response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal pipelines.yaml: %s", err.Error()), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(response)
}

// NewRouter creates and returns a new APIRouter.
func NewRouter(c git.SCM) *APIRouter {
	api := &APIRouter{Router: httprouter.New(), scmClient: c}
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
