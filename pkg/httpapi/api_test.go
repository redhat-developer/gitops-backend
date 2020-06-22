package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bigkevmcd/gitops-backend/pkg/git"
	"github.com/bigkevmcd/gitops-backend/test"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

func TestGetPipelines(t *testing.T) {
	c := newClient()
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
	ts := httptest.NewTLSServer(AuthenticationMiddleware(NewRouter(c)))
	t.Cleanup(ts.Close)
	pipelinesURL := "https://github.com/example/gitops.git"

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?url=%s", ts.URL, pipelinesURL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":     "taxi",
				"repo_url": "https://example.com/demo/gitops.git",
			},
		},
	})
}

func TestGetPipelinesWithNoURL(t *testing.T) {
	c := newClient()
	ts := httptest.NewTLSServer(AuthenticationMiddleware(NewRouter(c)))
	t.Cleanup(ts.Close)

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines", ts.URL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusBadRequest, "missing parameter 'url'")
}

func TestGetPipelinesWithBadURL(t *testing.T) {
	c := newClient()
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
	ts := httptest.NewTLSServer(AuthenticationMiddleware(NewRouter(c)))
	t.Cleanup(ts.Close)

	req := makeClientRequest(t, "Bearer testing", fmt.Sprintf("%s/pipelines?url=%%%%test.html", ts.URL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusBadRequest, "missing parameter 'url'")
}

func TestGetPipelinesWithPrivateRepoAndNoCredentials(t *testing.T) {
	t.Skip()
}

func TestGetPipelinesWithNoAuthorizationHeader(t *testing.T) {
	c := newClient()
	ts := httptest.NewTLSServer(AuthenticationMiddleware(NewRouter(c)))
	t.Cleanup(ts.Close)

	req := makeClientRequest(t, "", fmt.Sprintf("%s/pipelines", ts.URL))
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPError(t, res, http.StatusForbidden, "Authentication required")
}

// TODO: assert the content-type.
func assertJSONResponse(t *testing.T, res *http.Response, want map[string]interface{}) {
	t.Helper()
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		errMsg, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("didn't get a successful response: %v (%s)", res.StatusCode, errMsg)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
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

func TestPipelinesToAppsResponse(t *testing.T) {
	raw := parseYAMLToConfig(t, "testdata/pipelines.yaml")

	apps := pipelinesToAppsResponse(raw)

	want := &appsResponse{
		Apps: []appResponse{
			{
				Name: "taxi", RepoURL: "https://example.com/demo/gitops.git",
			},
		},
	}
	if diff := cmp.Diff(want, apps); diff != "" {
		t.Fatalf("failed to parse:\n%s", diff)
	}
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
