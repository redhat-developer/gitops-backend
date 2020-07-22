package parser

import (
	"sort"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/gvk"
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

func resKey(r *resource.Resource) string {
	return strings.Join([]string{r.Name, r.Namespace, r.Group, r.Kind, r.Version}, "-")
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
	sort.SliceStable(res, func(i, j int) bool { return resKey(res[i]) < resKey(res[j]) })

	want := []*resource.Resource{
		{
			Group: "apps", Version: "v1", Kind: "Deployment", Name: "go-demo-http",
			Labels: map[string]string{
				nameLabel:   "go-demo",
				partOfLabel: "go-demo",
			},
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
		},
	}
	sort.SliceStable(want, func(i, j int) bool { return resKey(want[i]) < resKey(want[j]) })
	assertCmp(t, want, res, "failed to match parsed resources")
}

func TestExtractResource(t *testing.T) {
	redisMap := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				nameLabel:   "redis",
				partOfLabel: "go-demo",
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
		Labels:    map[string]string{nameLabel: "redis", partOfLabel: "go-demo"},
	}
	assertCmp(t, want, svc, "failed to match resource")
}

func assertCmp(t *testing.T, want, got interface{}, msg string) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf(msg+":\n%s", diff)
	}
}
