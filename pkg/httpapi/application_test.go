package httpapi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

const (
	nameLabel   = "app.kubernetes.io/name"
	partOfLabel = "app.kubernetes.io/part-of"
)

func TestParseServicesFromResources(t *testing.T) {
	res := []*resource.Resource{
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
	}

	env := &environment{
		Name:    "test-env",
		Cluster: "https://cluster.local",
		Apps: []*application{
			{Name: "my-app"},
		},
	}

	svcs := parseServicesFromResources(env, res)
	want := []service{}
	if diff := cmp.Diff(want, svcs); diff != "" {
		t.Fatalf("parseServicesFromResources got\n%s", diff)
	}

}
