package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	argoV1aplha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	gogit "github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/gitops-backend/pkg/git"
	"github.com/redhat-developer/gitops-backend/pkg/parser"
	"github.com/redhat-developer/gitops-backend/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

const (
	testRef = "7638417db6d59f3c431d3e1f261cc637155684cd"
)

func TestGetPipelines(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", "HEAD", "testdata/pipelines.yaml")
	pipelinesURL := "https://github.com/example/gitops.git"

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?url=%s", ts.URL, pipelinesURL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":         "taxi",
				"repo_url":     "https://example.com/demo/gitops.git",
				"environments": []interface{}{"dev"},
			},
		},
	})
}

func TestGetPipelinesWithASpecificRef(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", testRef, "testdata/pipelines.yaml")
	pipelinesURL := fmt.Sprintf("https://github.com/example/gitops.git?ref=%s", testRef)

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?url=%s", ts.URL, pipelinesURL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":         "taxi",
				"repo_url":     "https://example.com/demo/gitops.git",
				"environments": []interface{}{"dev"},
			},
		},
	})
}

func TestGetPipelinesWithNoURL(t *testing.T) {
	ts, _ := makeServer(t)
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines", ts.URL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusBadRequest, "missing parameter 'url'")
}

func TestGetPipelinesWithBadURL(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", "main", "testdata/pipelines.yaml")
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?url=%%%%test.html", ts.URL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusBadRequest, "missing parameter 'url'")
}

func TestGetPipelinesWithNoAuthorizationHeader(t *testing.T) {
	ts, _ := makeServer(t)
	req := makeClientRequest(t, "", fmt.Sprintf("%s/pipelines", ts.URL))

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusForbidden, "Authentication required")
}

func TestGetPipelinesWithNamespaceAndNameInURL(t *testing.T) {
	secretRef := types.NamespacedName{
		Name:      "other-name",
		Namespace: "other-ns",
	}
	sg := &stubSecretGetter{
		testToken:     "test-token",
		testName:      secretRef,
		testAuthToken: "testing",
		testKey:       "token",
	}
	ts, c := makeServer(t, func(a *APIRouter) {
		a.secretGetter = sg
	})
	c.addContents("example/gitops", "pipelines.yaml", "HEAD", "testdata/pipelines.yaml")
	pipelinesURL := "https://github.com/example/gitops.git"
	options := url.Values{
		"url":        []string{pipelinesURL},
		"secretName": []string{"other-name"},
		"secretNS":   []string{"other-ns"},
	}
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?%s", ts.URL, options.Encode()))

	res, err := ts.Client().Do(req)

	if err != nil {
		t.Fatal(err)
	}
	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":         "taxi",
				"repo_url":     "https://example.com/demo/gitops.git",
				"environments": []interface{}{"dev"},
			},
		},
	})
}

func TestGetPipelinesWithUnknownSecret(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", "main", "testdata/pipelines.yaml")
	pipelinesURL := "https://github.com/example/gitops.git"
	options := url.Values{
		"url":        []string{pipelinesURL},
		"secretName": []string{"other-name"},
		"secretNS":   []string{"other-ns"},
	}
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?%s", ts.URL, options.Encode()))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertErrorResponse(t, res, http.StatusBadRequest, "unable to authenticate request")
}

func TestGetPipelineApplication(t *testing.T) {
	testResource := &parser.Resource{
		Group:     "",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment",
		Namespace: "test-ns",
		Labels: map[string]string{
			nameLabel: "gitops-demo",
		},
	}

	ts, c := makeServer(t, func(a *APIRouter) {
		a.resourceParser = stubResourceParser(testResource)
	})
	c.addContents("example/gitops", "pipelines.yaml", "HEAD", "testdata/pipelines.yaml")
	pipelinesURL := "https://github.com/example/gitops.git"
	options := url.Values{
		"url": []string{pipelinesURL},
	}
	req := makeClientRequest(t, "Bearer testing",
		fmt.Sprintf("%s/environments/%s/application/%s?%s", ts.URL, "dev", "taxi", options.Encode()))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"environment": "dev",
		"cluster":     "https://dev.testing.svc",
		"services": []interface{}{
			map[string]interface{}{
				"name": "gitops-demo",
				"resources": []interface{}{
					map[string]interface{}{
						"group":     "",
						"kind":      "Deployment",
						"name":      "test-deployment",
						"namespace": "test-ns",
						"version":   "v1",
					},
				},
				"source": map[string]interface{}{
					"type": "example.com",
					"url":  "https://example.com/demo/gitops-demo.git",
				},
			},
		},
	})
}

func TestGetPipelineApplicationWithRef(t *testing.T) {
	testResource := &parser.Resource{
		Group:     "",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment",
		Namespace: "test-ns",
		Labels: map[string]string{
			nameLabel: "gitops-demo",
		},
	}

	ts, c := makeServer(t, func(a *APIRouter) {
		a.resourceParser = stubResourceParser(testResource)
	})
	c.addContents("example/gitops", "pipelines.yaml", testRef, "testdata/pipelines.yaml")
	pipelinesURL := fmt.Sprintf("https://github.com/example/gitops.git?ref=%s", testRef)
	options := url.Values{
		"url": []string{pipelinesURL},
	}
	req := makeClientRequest(t, "Bearer testing",
		fmt.Sprintf("%s/environments/%s/application/%s?%s", ts.URL, "dev", "taxi", options.Encode()))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"environment": "dev",
		"cluster":     "https://dev.testing.svc",
		"services": []interface{}{
			map[string]interface{}{
				"name": "gitops-demo",
				"resources": []interface{}{
					map[string]interface{}{
						"group":     "",
						"kind":      "Deployment",
						"name":      "test-deployment",
						"namespace": "test-ns",
						"version":   "v1",
					},
				},
				"source": map[string]interface{}{
					"type": "example.com",
					"url":  "https://example.com/demo/gitops-demo.git",
				},
			},
		},
	})
}

func TestParseURL(t *testing.T) {
	urlTests := []struct {
		u        string
		wantRepo string
		wantErr  string
	}{
		{"https://github.com/example/gitops.git?ref=main", "example/gitops", ""},
		{"%%foo.html", "", "invalid URL escape"},
		{"https://github.com/example/testing.git", "example/testing", ""},
	}

	for _, tt := range urlTests {
		repo, got, err := parseURL(tt.u)
		if !test.MatchError(t, tt.wantErr, err) {
			t.Errorf("got an unexpected error: %v", err)
			continue
		}
		if err == nil {
			want, err := url.Parse(tt.u)
			assertNoError(t, err)
			if got.String() != want.String() {
				t.Errorf("Parsed URL mismatch: got %v, want %v", got.String(), want.String())
			}
		}
		if repo != tt.wantRepo {
			t.Errorf("repo got %s, want %s", repo, tt.wantRepo)
		}
	}
}

func TestListApplications(t *testing.T) {
	err := argoV1aplha1.AddToScheme(scheme.Scheme)
	if err != nil {
		t.Fatal(err)
	}

	builder := fake.NewClientBuilder()
	kc := builder.Build()

	ts, _ := makeServer(t, func(router *APIRouter) {
		router.k8sClient = kc
	})

	var createOptions []ctrlclient.CreateOption
	app, _ := testArgoApplication("testdata/application.yaml")
	err = kc.Create(context.TODO(), app, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	url := "https://github.com/test-repo/gitops.git?ref=HEAD"
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/applications?url=%s", ts.URL, url))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":          "test-app",
				"repo_url":      "https://github.com/test-repo/gitops.git",
				"environments":  []interface{}{"dev"},
				"sync_status":   []interface{}{"Synced"},
				"last_deployed": []interface{}{time.Date(2021, time.Month(5), 15, 2, 12, 13, 0, time.UTC).Local().String()},
			},
		},
	})
}

func TestListApplicationsWIthTwoApps(t *testing.T) {
	err := argoV1aplha1.AddToScheme(scheme.Scheme)
	if err != nil {
		t.Fatal(err)
	}

	builder := fake.NewClientBuilder()
	kc := builder.Build()

	ts, _ := makeServer(t, func(router *APIRouter) {
		router.k8sClient = kc
	})

	var createOptions []ctrlclient.CreateOption
	app, _ := testArgoApplication("testdata/application.yaml")
	err = kc.Create(context.TODO(), app, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	app2, _ := testArgoApplication("testdata/application2.yaml")
	err = kc.Create(context.TODO(), app2, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	url := "https://github.com/test-repo/gitops.git?ref=HEAD"
	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/applications?url=%s", ts.URL, url))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":         "test-app",
				"repo_url":     "https://github.com/test-repo/gitops.git",
				"environments": []interface{}{"dev", "production"},
				"sync_status":  []interface{}{"Synced", "OutOfSync"},
				"last_deployed": []interface{}{time.Date(2021, time.Month(5), 15, 2, 12, 13, 0, time.UTC).Local().String(),
					time.Date(2021, time.Month(5), 16, 1, 10, 35, 0, time.UTC).Local().String(),
				},
			},
		},
	})
}

func TestListApplications_badURL(t *testing.T) {
	builder := fake.NewClientBuilder()
	kc := builder.Build()

	ts, _ := makeServer(t, func(router *APIRouter) {
		router.k8sClient = kc
	})

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/applications", ts.URL))
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertHTTPError(t, resp, http.StatusBadRequest, "please provide a valid GitOps repo URL")
}

func TestGetApplicationDetails(t *testing.T) {
	err := argoV1aplha1.AddToScheme(scheme.Scheme)
	if err != nil {
		t.Fatal(err)
	}

	builder := fake.NewClientBuilder()
	kc := builder.Build()

	ts, _ := makeServer(t, func(router *APIRouter) {
		router.k8sClient = kc
	})

	// create test ArgoCD Server to handle http requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.String()
		if strings.Contains(u, "session") {
			m := map[string]string{
				"token": "testing",
			}
			marshalResponse(w, m)
		} else if strings.Contains(u, "metadata") {
			m := map[string]string{
				"author":  "test",
				"message": "testMessage",
			}
			marshalResponse(w, m)
		}
	}))
	defer server.Close()
	tmp := baseURL
	baseURL = server.URL

	var createOptions []ctrlclient.CreateOption
	app, _ := testArgoApplication("testdata/application.yaml")
	// create argocd test-app
	err = kc.Create(context.TODO(), app, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	// create argocd instance creds secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultArgoCDInstance + "-cluster",
			Namespace: defaultArgocdNamespace,
		},
		Data: map[string][]byte{
			"admin.password": []byte("abc"),
		},
	}
	err = kc.Create(context.TODO(), secret, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	options := url.Values{
		"url": []string{"https://github.com/test-repo/gitops.git"},
	}
	req := makeClientRequest(t, "Bearer testing",
		fmt.Sprintf("%s/environment/%s/application/%s?%s", ts.URL, "dev", "test-app", options.Encode()))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"cluster":      "https://kubernetes.default.svc",
		"environment":  "dev",
		"status":       "Synced",
		"lastDeployed": time.Date(2021, time.Month(5), 15, 2, 12, 13, 0, time.UTC).Local().String(),
		"revision": map[string]interface{}{
			"author":   "test",
			"message":  "testMessage",
			"revision": "123456789",
		},
		"deployments": []interface{}{
			map[string]interface{}{"health": string("Healthy"), "name": string("taxi"), "status": string("Synced")},
		},
		"secrets": []interface{}{
			map[string]interface{}{
				"health": string("Missing"),
				"name":   string("testsecret"),
				"status": string("OutOfSync"),
			},
		},
		"services": []interface{}{
			map[string]interface{}{"health": string("Healthy"), "name": string("taxi"), "status": string("Synced")},
		},
		"routes": []interface{}{
			map[string]interface{}{"name": string("taxi"), "status": string("Synced")},
		},
		"clusterRoleBindings": []interface{}{
			map[string]interface{}{"name": string("pipelines-service-role-binding"), "status": string("Synced")},
		},
		"clusterRoles": []interface{}{
			map[string]interface{}{"name": string("pipelines-clusterrole"), "status": string("Synced")},
		},
		"roleBindings": []interface{}{
			map[string]interface{}{"name": string("argocd-admin"), "status": string("Synced")},
		},
	})

	//reset BaseURL
	baseURL = tmp
}

func TestGetApplicationHistory(t *testing.T) {
	err := argoV1aplha1.AddToScheme(scheme.Scheme)
	if err != nil {
		t.Fatal(err)
	}

	builder := fake.NewClientBuilder()
	kc := builder.Build()

	ts, _ := makeServer(t, func(router *APIRouter) {
		router.k8sClient = kc
	})

	// create test ArgoCD Server to handle http requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.String()
		if strings.Contains(u, "session") {
			m := map[string]string{
				"token": "testing",
			}
			marshalResponse(w, m)
		} else if strings.Contains(u, "metadata") {
			m := map[string]string{
				"author":  "test",
				"message": "testMessage",
			}
			marshalResponse(w, m)
		}
	}))
	defer server.Close()
	tmp := baseURL
	baseURL = server.URL

	var createOptions []ctrlclient.CreateOption
	app, _ := testArgoApplication("testdata/application3.yaml")
	// create argocd test-app
	err = kc.Create(context.TODO(), app, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	// create argocd instance creds secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultArgoCDInstance + "-cluster",
			Namespace: defaultArgocdNamespace,
		},
		Data: map[string][]byte{
			"admin.password": []byte("abc"),
		},
	}
	err = kc.Create(context.TODO(), secret, createOptions...)
	if err != nil {
		t.Fatal(err)
	}

	options := url.Values{
		"url": []string{"https://github.com/test-repo/gitops.git"},
	}
	req := makeClientRequest(t, "Bearer testing",
		fmt.Sprintf("%s/history/environment/%s/application/%s?%s", ts.URL, "dev", "app-taxi", options.Encode()))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("history res is ", res)

	want := []envHistory{
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "a0c7298faead28f7f60a5106afbb18882ad220a7",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 22, 17, 11, 29, 0, time.UTC).Local().String(),
		},
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "3f6965bd65d9294b8fec5d6e2dc3dad08e33a8fe",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 21, 14, 17, 49, 0, time.UTC).Local().String(),
		},
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "e5585fcf22366e2d066e0936cbd8a0508756d02d",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 21, 14, 16, 51, 0, time.UTC).Local().String(),
		},
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "3f6965bd65d9294b8fec5d6e2dc3dad08e33a8fe",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 21, 14, 16, 50, 0, time.UTC).Local().String(),
		},
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "e5585fcf22366e2d066e0936cbd8a0508756d02d",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 21, 14, 14, 27, 0, time.UTC).Local().String(),
		},
		{
			Author:      "test",
			Message:     "testMessage",
			Revision:    "e5585fcf22366e2d066e0936cbd8a0508756d02d",
			Environment: "dev",
			RepoUrl:     "https://github.com/test-repo/gitops.git",
			DeployedAt:  time.Date(2022, time.Month(4), 19, 18, 19, 52, 0, time.UTC).Local().String(),
		},
	}
	assertJSONResponseHistory(t, res, want)

	//reset BaseURL
	baseURL = tmp
}

func testArgoApplication(appCr string) (*argoV1aplha1.Application, error) {
	applicationYaml, _ := ioutil.ReadFile(appCr)
	app := &argoV1aplha1.Application{}
	err := yaml.Unmarshal(applicationYaml, app)
	if err != nil {
		return nil, err
	}

	return app, err
}

func newClient() *stubClient {
	return &stubClient{files: make(map[string]string)}
}

type stubClient struct {
	files map[string]string
}

func (s stubClient) FileContents(ctx context.Context, repo, path, ref string) ([]byte, error) {
	f, ok := s.files[key(repo, path, ref)]
	if !ok {
		return nil, git.SCMError{Status: http.StatusNotFound}
	}
	return ioutil.ReadFile(f)
}

func (s *stubClient) addContents(repo, path, ref, filename string) {
	s.files[key(repo, path, ref)] = filename
}

func parseYAMLToConfig(t *testing.T, path string) *config {
	t.Helper()
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	response := &config{}
	err = yaml.Unmarshal(b, &response)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func key(s ...string) string {
	return strings.Join(s, "#")
}

func makeClientRequest(t *testing.T, token, path string) *http.Request {
	r, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set(authHeader, token)
	return r
}

type routerOptionFunc func(*APIRouter)

func makeServer(t *testing.T, opts ...routerOptionFunc) (*httptest.Server, *stubClient) {
	sg := &stubSecretGetter{
		testToken:     "test-token",
		testName:      DefaultSecretRef,
		testAuthToken: "testing",
		testKey:       "token",
	}
	sf := &stubClientFactory{client: newClient()}
	var kc ctrlclient.Client
	router := NewRouter(sf, sg, kc)
	for _, o := range opts {
		o(router)
	}

	ts := httptest.NewTLSServer(AuthenticationMiddleware(router))
	t.Cleanup(ts.Close)
	return ts, sf.client
}

func readBody(t *testing.T, res *http.Response) []byte {
	t.Helper()
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		errMsg, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("didn't get a successful response: %v (%s)", res.StatusCode, strings.TrimSpace(string(errMsg)))
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if h := res.Header.Get("Content-Type"); h != "application/json" {
		t.Fatalf("wanted 'application/json' got %s", h)
	}
	if h := res.Header.Get("Access-Control-Allow-Origin"); h != "*" {
		t.Fatalf("wanted '*' got %s", h)
	}
	return b
}

func assertJSONResponse(t *testing.T, res *http.Response, want map[string]interface{}) {
	b := readBody(t, res)
	got := map[string]interface{}{}

	err := json.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("failed to parse %s: %s", b, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("JSON response failed:\n%s", diff)
	}
}

func assertJSONResponseHistory(t *testing.T, res *http.Response, want []envHistory) {
	b := readBody(t, res)
	got := make([]envHistory, 0)
	err := json.Unmarshal(b, &got)
	if err != nil {
		t.Fatalf("failed to parse %s: %s", b, err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("JSON response failed:\n%s", diff)
	}
}

func assertErrorResponse(t *testing.T, res *http.Response, status int, want string) {
	t.Helper()
	if res.StatusCode != status {
		defer res.Body.Close()
		errMsg, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("status code didn't match: %v (%s)", res.StatusCode, strings.TrimSpace(string(errMsg)))
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

type stubSecretGetter struct {
	testAuthToken string
	testToken     string
	testName      types.NamespacedName
	testKey       string
}

func (f *stubSecretGetter) SecretToken(ctx context.Context, authToken string, id types.NamespacedName, key string) (string, error) {
	if id == f.testName && authToken == f.testAuthToken && key == f.testKey {
		return f.testToken, nil
	}
	return "", errors.New("failed to get a secret token")
}

type stubClientFactory struct {
	client *stubClient
}

func (s stubClientFactory) Create(url, token string) (git.SCM, error) {
	// TODO: this should match on the URL/token combo.
	return s.client, nil
}

func stubResourceParser(r ...*parser.Resource) parser.ResourceParser {
	return func(path string, opts *gogit.CloneOptions) ([]*parser.Resource, error) {
		return r, nil
	}
}
