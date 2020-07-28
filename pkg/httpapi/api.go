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
	"github.com/rhd-gitops-examples/gitops-backend/pkg/parser"
)

// DefaultSecretRef is the name looked up if none is provided in the URL.
var DefaultSecretRef = types.NamespacedName{
	Name:      "pipelines-app-gitops",
	Namespace: "pipelines-app-delivery",
}

const defaultRef = "master"

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
	gitClientFactory git.ClientFactory
	secretGetter     secrets.SecretGetter
	secretRef        types.NamespacedName
	resourceParser   parser.ResourceParser
	driverIdentifier git.DriverIdentifier
}

// NewRouter creates and returns a new APIRouter.
func NewRouter(c git.ClientFactory, s secrets.SecretGetter) *APIRouter {
	api := &APIRouter{
		Router:           httprouter.New(),
		gitClientFactory: c,
		secretGetter:     s,
		secretRef:        DefaultSecretRef,
		resourceParser:   parser.ParseFromGit,
	}
	api.HandlerFunc(http.MethodGet, "/pipelines", api.GetPipelines)
	api.HandlerFunc(http.MethodGet, "/environments/:env/application/:app", api.GetApplication)
	return api
}

// GetPipelines fetches and returns the pipeline body.
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

	token, err := a.getAuthToken(r.Context(), r)
	if err != nil {
		log.Printf("ERROR: failed to get an authentication token: %s", err)
		http.Error(w, "unable to authenticate request", http.StatusBadRequest)
		return
	}
	client, err := a.getAuthenticatedGitClient(urlToFetch, token)
	if err != nil {
		log.Printf("ERROR: failed to get an authenticated client: %s", err)
		http.Error(w, "unable to authenticate request", http.StatusBadRequest)
		return
	}

	// TODO: don't send back the error directly.
	//
	// Add a "not found" error that can be returned, otherwise it's a
	// StatusInternalServerError.
	log.Println("got an authenticated client")
	body, err := client.FileContents(r.Context(), repo, "pipelines.yaml", refFromQuery(r.URL.Query()))
	if err != nil {
		log.Printf("ERROR: failed to get file contents for repo %#v: %s", repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pipelines := &config{}
	err = yaml.Unmarshal(body, &pipelines)
	if err != nil {
		log.Printf("ERROR: failed to unmarshal body: %s", err)
		http.Error(w, fmt.Sprintf("failed to unmarshal pipelines.yaml: %s", err.Error()), http.StatusBadRequest)
		return
	}
	marshalResponse(w, pipelinesToAppsResponse(pipelines))
}

// GetApplication fetches an application within a specific environment.
//
// Expects the
func (a *APIRouter) GetApplication(w http.ResponseWriter, r *http.Request) {
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

	token, err := a.getAuthToken(r.Context(), r)
	if err != nil {
		log.Printf("ERROR: failed to get an authentication token: %s", err)
		http.Error(w, "unable to authenticate request", http.StatusBadRequest)
		return
	}
	client, err := a.getAuthenticatedGitClient(urlToFetch, token)
	if err != nil {
		log.Printf("ERROR: failed to get an authenticated client: %s", err)
		http.Error(w, "unable to authenticate request", http.StatusBadRequest)
		return
	}

	// TODO: don't send back the error directly.
	//
	// Add a "not found" error that can be returned, otherwise it's a
	// StatusInternalServerError.
	body, err := client.FileContents(r.Context(), repo, "pipelines.yaml", refFromQuery(r.URL.Query()))
	if err != nil {
		log.Printf("ERROR: failed to get file contents for repo %#v: %s", repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pipelines := &config{}
	err = yaml.Unmarshal(body, &pipelines)
	if err != nil {
		log.Printf("ERROR: failed to unmarshal body: %s", err)
		http.Error(w, fmt.Sprintf("failed to unmarshal pipelines.yaml: %s", err.Error()), http.StatusBadRequest)
		return
	}
	params := httprouter.ParamsFromContext(r.Context())
	appEnvironments, err := a.environmentApplication(token, pipelines, params.ByName("env"), params.ByName("app"))
	if err != nil {
		log.Printf("ERROR: failed to get application data: %s", err)
		http.Error(w, "failed to extract data", http.StatusBadRequest)
		return
	}
	marshalResponse(w, appEnvironments)
}

func (a *APIRouter) getAuthToken(ctx context.Context, req *http.Request) (string, error) {
	token := AuthToken(ctx)
	secret, ok := secretRefFromQuery(req.URL.Query())
	if !ok {
		secret = a.secretRef
	}
	// TODO: this should be using a logger implementation.
	log.Printf("using secret from %#v", secret)
	token, err := a.secretGetter.SecretToken(ctx, token, secret, "token")
	if err != nil {
		return "", err
	}
	return token, nil
}
func (a *APIRouter) getAuthenticatedGitClient(fetchURL, token string) (git.SCM, error) {
	return a.gitClientFactory.Create(fetchURL, token)
}

func parseURL(s string) (string, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("failed to parse %#v: %w", s, err)
	}
	return strings.TrimLeft(strings.Trim(parsed.Path, ".git"), "/"), nil
}

func secretRefFromQuery(v url.Values) (types.NamespacedName, bool) {
	ns := v.Get("secretNS")
	name := v.Get("secretName")
	if ns != "" && name != "" {
		return types.NamespacedName{
			Name:      name,
			Namespace: ns,
		}, true
	}
	return types.NamespacedName{}, false
}

func refFromQuery(v url.Values) string {
	if ref := v.Get("ref"); ref != "" {
		return ref
	}
	return defaultRef
}

func marshalResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(v)
}
