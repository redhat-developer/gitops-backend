package parser

import (
	"sort"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/gvk"
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

func TestParseFromGit(t *testing.T) {
	res, err := ParseFromGit(
		"pkg/parser/testdata/go-demo",
		&git.CloneOptions{
			URL:   "../..",
			Depth: 1,
		})
	if err != nil {
		t.Fatal(err)
	}
	sort.SliceStable(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	want := []*resource.Resource{
		{Group: "apps", Version: "v1", Kind: "Deployment", Name: "go-demo-http"},
		{Version: "v1", Kind: "Service", Name: "go-demo-http"},
		{Version: "v1", Kind: "ConfigMap", Name: "go-demo-config"},
		{Version: "v1", Kind: "Service", Name: "redis"},
		{Group: "apps", Version: "v1", Kind: "Deployment", Name: "redis"},
	}
	sort.SliceStable(want, func(i, j int) bool { return want[i].Name < want[j].Name })
	assertCmp(t, want, res, "failed to match parsed resources")
}

func TestExtractResource(t *testing.T) {
	redisMap := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"app.kubernetes.io/name":    "redis",
				"app.kubernetes.io/part-of": "go-demo",
			},
			"name":      "redis",
			"namespace": "test-env",
		},
	}

	svc := extractResource(gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}, redisMap)
	want := &resource.Resource{
		Name:      "redis",
		Namespace: "test-env",
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
	}
	assertCmp(t, want, svc, "failed to match resource")
}

func assertCmp(t *testing.T, want, got interface{}, msg string) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf(msg+":\n%s", diff)
	}
}
