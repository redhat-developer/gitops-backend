package httpapi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPipelinesToAppsResponse(t *testing.T) {
	raw := parseYAMLToConfig(t, "testdata/pipelines.yaml")

	apps := pipelinesToAppsResponse(raw)

	want := &appsResponse{
		Apps: []appResponse{
			{
				Name: "taxi", RepoURL: "https://example.com/demo/gitops.git",
				Environments: []string{"tst-dev", "tst-stage"},
			},
		},
	}
	if diff := cmp.Diff(want, apps); diff != "" {
		t.Fatalf("failed to parse:\n%s", diff)
	}
}
