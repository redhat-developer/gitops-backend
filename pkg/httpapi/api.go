package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/httpapi/secrets"
)

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
	gitClientFactory GitClientFactory
	secretGetter     secrets.SecretGetter
	// TODO: replace this with a way to get it from the request.
	SecretRef types.NamespacedName
}

// GePipelines fetches and returns the pipeline body.
func (a *APIRouter) GetPipelines(w http.ResponseWriter, r *http.Request) {
	urlToFetch := r.URL.Query().Get("url")
	if urlToFetch == "" {
		log.Println("ERROR: could not get url from request")
		http.Error(w, "missing parameter 'url'", http.StatusBadRequest)
		return
	}

	// TODO: replace this with logr or sugar.
	log.Printf("urlToFetch = %#v\n", urlToFetch)
	repo, err := parseURL(urlToFetch)
	if err != nil {
		log.Printf("ERROR: failed to parse the URL: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: don't send back the error directly, it could contain technical
	// details.
	client, err := a.getAuthenticatedGitClient(r.Context(), urlToFetch)
	if err != nil {
		log.Println("ERROR: failed to get an authenticated client")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: don't send back the error directly.
	//
	// Add a "not found" error that can be returned, otherwise it's a
	// StatusInternalServerError.
	log.Println("got an authenticated client")
	body, err := client.FileContents(r.Context(), repo, "pipelines.yaml", "master")
	if err != nil {
		log.Printf("ERROR: failed to get file contents for repo %#v: %s", repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pipelines := &config{}
	err = yaml.Unmarshal(body, &pipelines)
	if err != nil {
		log.Printf("ERROR: failed to unmarshal body %s", err)
		http.Error(w, fmt.Sprintf("failed to unmarshal pipelines.yaml: %s", err.Error()), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(pipelinesToAppsResponse(pipelines))
}

func (a *APIRouter) getAuthenticatedGitClient(ctx context.Context, url string) (git.SCM, error) {
	token, _ := AuthToken(ctx)
	token, err := a.secretGetter.SecretToken(ctx, token, a.SecretRef)
	if err != nil {
		return nil, err
	}
	return a.gitClientFactory.Create(url, token)
}

// NewRouter creates and returns a new APIRouter.
func NewRouter(c GitClientFactory, s secrets.SecretGetter) *APIRouter {
	api := &APIRouter{Router: httprouter.New(), gitClientFactory: c, secretGetter: s}
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
