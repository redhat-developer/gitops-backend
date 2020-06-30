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

	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/git"
	"github.com/rhd-gitops-examples/gitops-backend/test"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

func TestGetPipelines(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
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
				"environments": []interface{}{"tst-dev", "tst-stage"},
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
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
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
	}
	ts, c := makeServer(t, func(a *APIRouter) {
		a.secretGetter = sg
	})
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
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
				"environments": []interface{}{"tst-dev", "tst-stage"},
			},
		},
	})
}

func TestGetPipelinesWithUnknownSecret(t *testing.T) {
	ts, c := makeServer(t)
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
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

func TestParseURL(t *testing.T) {
	urlTests := []struct {
		u        string
		wantRepo string
		wantErr  string
	}{
		{"https://github.com/example/gitops.git", "example/gitops", ""},
		{"%%foo.html", "", "invalid URL escape"},
	}

	for _, tt := range urlTests {
		repo, err := parseURL(tt.u)
		if !test.MatchError(t, tt.wantErr, err) {
			t.Errorf("got an unexpected error: %v", err)
			continue
		}
		if repo != tt.wantRepo {
			t.Errorf("repo got %s, want %s", repo, tt.wantRepo)
		}
	}
}

func TestSecretRefFromQuery(t *testing.T) {
	t.Skip()
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
	}
	sf := &stubClientFactory{client: newClient()}
	router := NewRouter(sf, sg)
	for _, o := range opts {
		o(router)
	}

	ts := httptest.NewTLSServer(AuthenticationMiddleware(router))
	t.Cleanup(ts.Close)
	return ts, sf.client
}

func assertJSONResponse(t *testing.T, res *http.Response, want map[string]interface{}) {
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
	got := map[string]interface{}{}
	err = json.Unmarshal(b, &got)
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
}

func (f *stubSecretGetter) SecretToken(ctx context.Context, authToken string, key types.NamespacedName) (string, error) {
	if key == f.testName && authToken == f.testAuthToken {
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
