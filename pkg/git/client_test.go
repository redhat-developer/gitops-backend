package git

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm/factory"

	"github.com/rhd-gitops-example/gitops-backend/pkg/metrics"
	"github.com/rhd-gitops-example/gitops-backend/test"
)

func TestFileContents(t *testing.T) {
	m := metrics.NewMock()
	as := makeAPIServer(t, "/api/v3/repos/Codertocat/Hello-World/contents/pipelines.yaml", "master", "testdata/content.json")
	defer as.Close()
	scmClient, err := factory.NewClient("github", as.URL, "", factory.Client(as.Client()))
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, m)

	body, err := client.FileContents(context.TODO(), "Codertocat/Hello-World", "pipelines.yaml", "master")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("testing service\n")
	if diff := cmp.Diff(want, body); diff != "" {
		t.Fatalf("got a different body back: %s\n", diff)
	}
	if m.APICalls != 1 {
		t.Fatalf("metrics count of API calls, got %d, want 1", m.APICalls)
	}
}

func TestFileContentsWithNotFoundResponse(t *testing.T) {
	m := metrics.NewMock()
	as := makeAPIServer(t, "/api/v3/repos/Codertocat/Hello-World/contents/pipelines.yaml", "master", "")
	defer as.Close()
	scmClient, err := factory.NewClient("github", as.URL, "", factory.Client(as.Client()))
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, m)

	_, err = client.FileContents(context.TODO(), "Codertocat/Hello-World", "pipelines.yaml", "master")
	if !IsNotFound(err) {
		t.Fatalf("failed with %#v", err)
	}
	if m.FailedAPICalls != 1 {
		t.Fatalf("metrics count of failed API calls, got %d, want 1", m.FailedAPICalls)
	}
}

func TestFileContentsUnableToConnect(t *testing.T) {
	m := metrics.NewMock()
	scmClient, err := factory.NewClient("github", "https://localhost:2000", "")
	if err != nil {
		t.Fatal(err)
	}
	client := New(scmClient, m)

	_, err = client.FileContents(context.TODO(), "Codertocat/Hello-World", "pipelines.yaml", "master")
	if !test.MatchError(t, "connection refused", err) {
		t.Fatal(err)
	}
	if m.FailedAPICalls != 1 {
		t.Fatalf("metrics count of failed API calls, got %d, want 1", m.FailedAPICalls)
	}
}

func makeAPIServer(t *testing.T, urlPath, ref, fixture string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if r.URL.Path != urlPath {
			http.NotFound(w, r)
			t.Fatalf("request path got %s, want %s", r.URL.Path, urlPath)
		}
		if ref != "" {
			if queryRef := r.URL.Query().Get("ref"); queryRef != ref {
				t.Fatalf("failed to match ref, got %s, want %s", queryRef, ref)
			}
		}
		if fixture == "" {
			http.NotFound(w, r)
			return
		}
		b, err := ioutil.ReadFile(fixture)
		if err != nil {
			t.Fatalf("failed to read %s: %s", fixture, err)
		}
		_, err = w.Write(b)
		if err != nil {
			t.Fatalf("failed to write: %s", err)
		}
	}))
}
