package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	argoV1aplha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/julienschmidt/httprouter"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/redhat-developer/gitops-backend/pkg/git"
	"github.com/redhat-developer/gitops-backend/pkg/httpapi/secrets"
	"github.com/redhat-developer/gitops-backend/pkg/parser"
)

// DefaultSecretRef is the name looked up if none is provided in the URL.
var DefaultSecretRef = types.NamespacedName{
	Name:      "pipelines-app-gitops",
	Namespace: "pipelines-app-delivery",
}

const defaultRef = "HEAD"

// APIRouter is an HTTP API for accessing app configurations.
type APIRouter struct {
	*httprouter.Router
	gitClientFactory git.ClientFactory
	secretGetter     secrets.SecretGetter
	secretRef        types.NamespacedName
	resourceParser   parser.ResourceParser
	k8sClient        ctrlclient.Client
}

// NewRouter creates and returns a new APIRouter.
func NewRouter(c git.ClientFactory, s secrets.SecretGetter, kc ctrlclient.Client) *APIRouter {
	api := &APIRouter{
		Router:           httprouter.New(),
		gitClientFactory: c,
		secretGetter:     s,
		secretRef:        DefaultSecretRef,
		resourceParser:   parser.ParseFromGit,
		k8sClient:        kc,
	}
	api.HandlerFunc(http.MethodGet, "/pipelines", api.GetPipelines)
	api.HandlerFunc(http.MethodGet, "/applications", api.ListApplications)
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
	repo, parsedRepo, err := parseURL(urlToFetch)
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
	body, err := client.FileContents(r.Context(), repo, "pipelines.yaml", refFromQuery(parsedRepo.Query()))
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
	repo, parsedRepo, err := parseURL(urlToFetch)
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
	body, err := client.FileContents(r.Context(), repo, "pipelines.yaml", refFromQuery(parsedRepo.Query()))
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

func (a *APIRouter) ListApplications(w http.ResponseWriter, r *http.Request) {
	repoURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if repoURL == "" {
		http.Error(w, "please provide a valid GitOps repo URL", http.StatusBadRequest)
		return
	}

	parsedRepoURL, err := url.Parse(repoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse URL, error: %v", err), http.StatusBadRequest)
		return
	}

	parsedRepoURL.RawQuery = ""

	appList := &argoV1aplha1.ApplicationList{}
	var listOptions []ctrlclient.ListOption

	listOptions = append(listOptions, ctrlclient.InNamespace(""))

	err = a.k8sClient.List(r.Context(), appList, listOptions...)
	if err != nil {
		log.Printf("ERROR: failed to get application list: %v", err)
		http.Error(w, fmt.Sprintf("failed to get list of application, err: %v", err), http.StatusBadRequest)
		return
	}

	apps := make([]*argoV1aplha1.Application, 0)
	for _, app := range appList.Items {
		apps = append(apps, app.DeepCopy())
	}

	marshalResponse(w, applicationsToAppsResponse(apps, parsedRepoURL.String()))
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

func parseURL(s string) (string, *url.URL, error) {
	parsed, err := url.Parse(s)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse %#v: %w", s, err)
	}
	return strings.TrimLeft(strings.TrimSuffix(parsed.Path, ".git"), "/"), parsed, nil
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
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %s", err)
	}
}
