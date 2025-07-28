package httpapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	argoV1aplha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/julienschmidt/httprouter"
	corev1 "k8s.io/api/core/v1"
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

const (
	defaultRef             = "HEAD"
	defaultArgoCDInstance  = "openshift-gitops"
	defaultArgocdNamespace = "openshift-gitops"
	kindService            = "Service"
	kindDeployment         = "Deployment"
	kindSecret             = "Secret"
	kindSealedSecret       = "SealedSecret"
	kindRoute              = "Route"
	kindRoleBinding        = "RoleBinding"
	kindClusterRole        = "ClusterRole"
	kindClusterRoleBinding = "ClusterRoleBinding"
)

var baseURL = fmt.Sprintf("https://%s-server.%s.svc.cluster.local", defaultArgoCDInstance, defaultArgocdNamespace)

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
	api.HandlerFunc(http.MethodGet, "/environment/:env/application/:app", api.GetApplicationDetails)
	api.HandlerFunc(http.MethodGet, "/history/environment/:env/application/:app", api.GetApplicationHistory)
	return api
}

type RevisionMeta struct {
	Author   string `json:"author"`
	Message  string `json:"message"`
	Revision string `json:"revision"`
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

func (a *APIRouter) GetApplicationHistory(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	envName, appName := params.ByName("env"), params.ByName("app")
	app := &argoV1aplha1.Application{}

	repoURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if repoURL == "" {
		log.Println("ERROR: please provide a valid GitOps repo URL")
		http.Error(w, "please provide a valid GitOps repo URL", http.StatusBadRequest)
		return
	}

	parsedRepoURL, err := url.Parse(repoURL)
	if err != nil {
		log.Printf("ERROR: failed to parse URL, error: %v", err)
		http.Error(w, fmt.Sprintf("failed to parse URL, error: %v", err), http.StatusBadRequest)
		return
	}

	parsedRepoURL.RawQuery = ""

	appList := &argoV1aplha1.ApplicationList{}
	var listOptions []ctrlclient.ListOption

	listOptions = append(listOptions, ctrlclient.InNamespace(""), ctrlclient.MatchingFields{
		"metadata.name": fmt.Sprintf("%s-%s", envName, appName),
	})

	err = a.k8sClient.List(r.Context(), appList, listOptions...)
	if err != nil {
		log.Printf("ERROR: failed to get application list: %v", err)
		http.Error(w, fmt.Sprintf("failed to get list of application, err: %v", err), http.StatusBadRequest)
		return
	}
	// At this point, if there are no apps as generated by KAM, then check if there are any
	// apps in general, where the app name is the environment name. Users aren't expected to use both
	// KAM generated apps and custom apps in their GitOps repo.
	if len(appList.Items) == 0 {
		listOptions = nil
		listOptions = append(listOptions, ctrlclient.InNamespace(""), ctrlclient.MatchingFields{
			"metadata.name": fmt.Sprintf("%s", envName),
		})
		err = a.k8sClient.List(r.Context(), appList, listOptions...)
		if err != nil {
			log.Printf("ERROR: failed to get application list: %v", err)
			http.Error(w, fmt.Sprintf("failed to get list of application, err: %v", err), http.StatusBadRequest)
			return
		}
	}

	for _, a := range appList.Items {
		if a.Spec.Source.RepoURL == parsedRepoURL.String() {
			app = &a
		}
	}

	if app == nil {
		log.Printf("ERROR: failed to get application %s: %v", appName, err)
		http.Error(w, fmt.Sprintf("failed to get the application %s, err: %v", appName, err), http.StatusBadRequest)
		return
	}

	var deployedTime, revision string
	hist := app.Status.History
	var historyList = make([]envHistory, 0)
	for _, h := range hist {
		revision = h.Revision
		t := h.DeployedAt
		if !t.IsZero() {
			deployedTime = t.String()
		}
		commitInfo, err := a.getCommitInfo(app.Name, revision)
		if err != nil {
			log.Printf("WARNING: failed to retrieve revision metadata for app %s: %v. The app might be unsynced.", appName, err)
		}
		hist := envHistory{
			Author:      commitInfo["author"],
			Message:     commitInfo["message"],
			Revision:    revision,
			RepoUrl:     h.Source.RepoURL,
			Environment: envName,
			DeployedAt:  deployedTime,
		}
		historyList = append([]envHistory{hist}, historyList...)
	}
	marshalResponse(w, historyList)
}

func (a *APIRouter) GetApplicationDetails(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	envName, appName := params.ByName("env"), params.ByName("app")
	app := &argoV1aplha1.Application{}
	var lastDeployed, revision string

	repoURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if repoURL == "" {
		log.Println("ERROR: please provide a valid GitOps repo URL")
		http.Error(w, "please provide a valid GitOps repo URL", http.StatusBadRequest)
		return
	}

	parsedRepoURL, err := url.Parse(repoURL)
	if err != nil {
		log.Printf("ERROR: failed to parse URL, error: %v", err)
		http.Error(w, fmt.Sprintf("failed to parse URL, error: %v", err), http.StatusBadRequest)
		return
	}

	parsedRepoURL.RawQuery = ""

	appList := &argoV1aplha1.ApplicationList{}
	var listOptions []ctrlclient.ListOption

	listOptions = append(listOptions, ctrlclient.InNamespace(""), ctrlclient.MatchingFields{
		"metadata.name": fmt.Sprintf("%s-%s", envName, appName),
	})

	err = a.k8sClient.List(r.Context(), appList, listOptions...)
	if err != nil {
		log.Printf("ERROR: failed to get application list: %v", err)
		http.Error(w, fmt.Sprintf("failed to get list of application, err: %v", err), http.StatusBadRequest)
		return
	}

	// At this point, if there are no apps as generated by KAM, then check if there are any
	// apps in general, where the app name is the environment name. Users aren't expected to use both
	// KAM generated apps and custom apps in their GitOps repo.
	// We should error out if we don't have the environments as applications
	if len(appList.Items) == 0 {
		listOptions = nil
		listOptions = append(listOptions, ctrlclient.InNamespace(""), ctrlclient.MatchingFields{
			"metadata.name": fmt.Sprintf("%s", envName),
		})
		err = a.k8sClient.List(r.Context(), appList, listOptions...)
		if err != nil {
			log.Printf("ERROR: failed to get application list: %v", err)
			http.Error(w, fmt.Sprintf("failed to get list of application, err: %v", err), http.StatusBadRequest)
			return
		}
	}

	for _, a := range appList.Items {
		if a.Spec.Source.RepoURL == parsedRepoURL.String() {
			app = &a
		}
	}

	if app == nil {
		log.Printf("ERROR: failed to get application %s: %v", appName, err)
		http.Error(w, fmt.Sprintf("failed to get the application %s, err: %v", appName, err), http.StatusBadRequest)
		return
	}

	if len(app.Status.History) > 0 {
		revision = app.Status.History[len(app.Status.History)-1].Revision
		t := app.Status.History[len(app.Status.History)-1].DeployedAt
		if !t.IsZero() {
			lastDeployed = t.String()
		}
	}

	commitInfo, err := a.getCommitInfo(app.Name, revision)
	if err != nil {
		log.Printf("Warning: failed to retrieve revision metadata for app %s: %v. The app might be unsynced", appName, err)
	}

	revisionMeta := RevisionMeta{
		Author:   strings.Split(commitInfo["author"], " ")[0],
		Message:  commitInfo["message"],
		Revision: revision,
	}
	var envResources = make(map[string][]envHealthResource)
	envResources[kindService] = make([]envHealthResource, 0)
	envResources[kindDeployment] = make([]envHealthResource, 0)
	envResources[kindSecret] = make([]envHealthResource, 0)
	envResources[kindRoute] = make([]envHealthResource, 0)
	envResources[kindRoleBinding] = make([]envHealthResource, 0)
	envResources[kindClusterRole] = make([]envHealthResource, 0)
	envResources[kindClusterRoleBinding] = make([]envHealthResource, 0)

	for _, aResource := range app.Status.Resources {
		switch aResource.Kind {
		case kindService:
			envResources[kindService] = append(envResources[kindService], envHealthResource{
				Name:   aResource.Name,
				Health: string(aResource.Health.Status),
				Status: string(aResource.Status),
			})
		case kindDeployment:
			envResources[kindDeployment] = append(envResources[kindDeployment], envHealthResource{
				Name:   aResource.Name,
				Health: string(aResource.Health.Status),
				Status: string(aResource.Status),
			})
		case kindSecret, kindSealedSecret:
			secretHealth := ""
			if aResource.Health != nil {
				secretHealth = string(aResource.Health.Status)
			}
			envResources[kindSecret] = append(envResources[kindSecret], envHealthResource{
				Name:   aResource.Name,
				Health: secretHealth,
				Status: string(aResource.Status),
			})
		case kindRoute:
			envResources[kindRoute] = append(envResources[kindRoute], envHealthResource{
				Name:   aResource.Name,
				Status: string(aResource.Status),
			})
		case kindRoleBinding:
			envResources[kindRoleBinding] = append(envResources[kindRoleBinding], envHealthResource{
				Name:   aResource.Name,
				Status: string(aResource.Status),
			})
		case kindClusterRole:
			envResources[kindClusterRole] = append(envResources[kindClusterRole], envHealthResource{
				Name:   aResource.Name,
				Status: string(aResource.Status),
			})
		case kindClusterRoleBinding:
			envResources[kindClusterRoleBinding] = append(envResources[kindClusterRoleBinding], envHealthResource{
				Name:   aResource.Name,
				Status: string(aResource.Status),
			})
		}
	}
	appEnv := map[string]interface{}{
		"environment":         app.Spec.Destination.Namespace,
		"cluster":             app.Spec.Destination.Server,
		"lastDeployed":        lastDeployed,
		"status":              app.Status.Sync.Status,
		"revision":            revisionMeta,
		"services":            envResources[kindService],
		"secrets":             envResources[kindSecret],
		"deployments":         envResources[kindDeployment],
		"routes":              envResources[kindRoute],
		"roleBindings":        envResources[kindRoleBinding],
		"clusterRoles":        envResources[kindClusterRole],
		"clusterRoleBindings": envResources[kindClusterRoleBinding],
	}

	marshalResponse(w, appEnv)
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

func (a *APIRouter) getCommitInfo(app, revision string) (map[string]string, error) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	argocdCreds := &corev1.Secret{}
	err := a.k8sClient.Get(context.TODO(),
		types.NamespacedName{
			Name:      defaultArgoCDInstance + "-cluster",
			Namespace: defaultArgocdNamespace,
		}, argocdCreds)
	if err != nil {
		return nil, err
	}

	payload := map[string]string{
		"username": "admin",
		"password": string(argocdCreds.Data["admin.password"]),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(fmt.Sprintf("%s/api/v1/session", baseURL),
		"application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	err = json.Unmarshal(bodyBytes, &m)
	if err != nil {
		return nil, err
	}

	if token, ok := m["token"]; !ok || token == "" {
		return nil, fmt.Errorf("failed to retrieve JWT from the api-server")
	}

	u := fmt.Sprintf("%s/api/v1/applications/%s/revisions/%s/metadata", baseURL, app, revision)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", m["token"]))
	req.Header.Add("content-type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bodyBytes, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
