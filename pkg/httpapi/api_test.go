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
)

func TestGetPipelines(t *testing.T) {
	c := newClient()
	c.addContents("example/gitops", "pipelines.yaml", "master", "testdata/pipelines.yaml")
	ts := httptest.NewTLSServer(NewRouter(c))
	t.Cleanup(ts.Close)
	pipelinesURL := "https://github.com/example/gitops.git"

	res, err := ts.Client().Get(fmt.Sprintf("%s/pipelines?url=%s", ts.URL, pipelinesURL))
	if err != nil {
		t.Fatal(err)
	}

	assertJSONResponse(t, res, map[string]interface{}{
		"environments": []interface{}{
			map[string]interface{}{
				"apps": []interface{}{
					map[string]interface{}{
						"name": "taxi",
						"services": []interface{}{
							map[string]interface{}{
								"name":       "taxi-svc",
								"source_url": "https://github.com/bigkevmcd/taxi.git",
								"webhook": map[string]interface{}{
									"secret": map[string]interface{}{
										"name":      "github-webhook-secret-taxi-svc",
										"namespace": "tst-cicd",
									},
								},
							},
						},
					},
				},
				"name": "tst-dev",
				"pipelines": map[string]interface{}{
					"integration": map[string]interface{}{
						"binding":  "github-pr-binding",
						"template": "app-ci-template"},
				},
			},
			map[string]interface{}{
				"name": "tst-stage",
			},
			map[string]interface{}{
				"cicd": true,
				"name": "tst-cicd",
			},
			map[string]interface{}{
				"argo": true,
				"name": "tst-argocd",
			},
		},
	})
}

func TestGetPipelinesWithNoURL(t *testing.T) {
	t.Skip()
}

func TestGetPipelinesWithBadURL(t *testing.T) {
	t.Skip()
}

// TODO: assert the content-type.
func assertJSONResponse(t *testing.T, res *http.Response, want map[string]interface{}) {
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

func key(s ...string) string {
	return strings.Join(s, "#")
}
