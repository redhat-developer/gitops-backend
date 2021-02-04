package parser

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"

	"github.com/redhat-developer/gitops-backend/test"
)

const (
	nameLabel   = "app.kubernetes.io/name"
	partOfLabel = "app.kubernetes.io/part-of"
)

func TestParseNoFile(t *testing.T) {
	res, err := ParseFromGit(
		"testdata",
		&git.CloneOptions{
			URL:   "../..",
			Depth: 1,
		})

	if res != nil {
		t.Errorf("did not expect to parse resources: %#v", res)
	}
	if err == nil {
		t.Fatal("expected to get an error")
	}
}

func resKey(r *Resource) string {
	return strings.Join([]string{r.Name, r.Namespace, r.Group, r.Kind, r.Version}, "-")
}

func TestParseFromGit(t *testing.T) {
	res, err := ParseFromGit(
		"pkg/parser/testdata/go-demo",
		test.MakeCloneOptions())
	if err != nil {
		t.Fatal(err)
	}
	sort.SliceStable(res, func(i, j int) bool { return resKey(res[i]) < resKey(res[j]) })

	want := []*Resource{
		{
			Group: "apps", Version: "v1", Kind: "Deployment", Name: "go-demo-http",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo",
			},
			Images: []string{"bigkevmcd/go-demo:876ecb3"},
		},
		{
			Version: "v1", Kind: "Service", Name: "go-demo-http",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo",
			},
		},
		{
			Version: "v1", Kind: "ConfigMap", Name: "go-demo-config",
			Labels: map[string]string{
				partOfLabel: "go-demo",
			},
		},
		{
			Version: "v1", Kind: "Service", Name: "redis",
			Labels: map[string]string{
				nameLabel:   "redis",
				partOfLabel: "go-demo",
			},
		},
		{
			Group: "apps", Version: "v1", Kind: "Deployment", Name: "redis",
			Labels: map[string]string{
				nameLabel:   "redis",
				partOfLabel: "go-demo",
			},
			Images: []string{"redis:6-alpine"},
		},
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "StatefulSet",
			Name:    "go-demo-web",
			Labels: map[string]string{
				partOfLabel: "go-demo",
				nameLabel:   "go-demo",
			},
			Images: []string{"bigkevmcd/go-demo-api:v0.0.1"},
		},
		{
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
			Name:    "demo-job",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo",
			},
			Images: []string{"bigkevmcd/go-demo:876ecb3"},
		},
		{
			Group:   "batch",
			Version: "v1beta1",
			Kind:    "CronJob",
			Name:    "hello",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo"},
			Images: []string{"alpine:latest"},
		},
		{
			Version: "v1",
			Group:   "apps.openshift.io",
			Kind:    "DeploymentConfig",
			Name:    "frontend",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo"},
			Images: []string{"demo/demo-config:v5"},
		},
	}
	sort.SliceStable(want, func(i, j int) bool { return resKey(want[i]) < resKey(want[j]) })
	assertCmp(t, want, res, "failed to match parsed resources")
}

func TestParseFromApplication(t *testing.T) {
	a := testReadApplication(t, "testdata/go_demo.yaml")
	r, err := ParseFromApplication(a)
	if err != nil {
		t.Fatal(err)
	}

	want := []*Resource{
		{Version: "v1", Kind: "ConfigMap", Name: "backend-demo-config", Namespace: "dev"},
		{Version: "v1", Kind: "Namespace", Name: "dev"},
		{Version: "v1", Kind: "Service", Name: "backend-demo-http", Namespace: "dev"},
		{Version: "v1", Kind: "Service", Name: "redis", Namespace: "dev"},
		{Group: "apps", Version: "v1", Kind: "Deployment", Name: "backend-demo-http", Namespace: "dev"},
		{Group: "apps", Version: "v1", Kind: "Deployment", Name: "redis", Namespace: "dev"},
	}
	if diff := cmp.Diff(want, r); diff != "" {
		t.Fatalf("parse failure:\n%s", diff)
	}
}

func testReadApplication(t *testing.T, s string) *appv1.Application {
	t.Helper()
	b, err := ioutil.ReadFile(s)
	if err != nil {
		t.Fatal(err)
	}
	a := &appv1.Application{}
	if err := yaml.Unmarshal(b, a); err != nil {
		t.Fatal(err)
	}
	return a
}

func assertCmp(t *testing.T, want, got interface{}, msg string) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf(msg+":\n%s", diff)
	}
}
