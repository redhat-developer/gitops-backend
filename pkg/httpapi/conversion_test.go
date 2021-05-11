package httpapi

import (
	"fmt"
	argoV1aplha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
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
				Environments: []string{"dev"},
			},
		},
	}
	if diff := cmp.Diff(want, apps); diff != "" {
		t.Fatalf("failed to parse:\n%s", diff)
	}
}

func TestApplicationsToAppsResponse(t *testing.T) {
	var apps []*argoV1aplha1.Application
	app, _ := testArgoApplication()
	apps = append(apps, app)

	want := &appsResponse{
		Apps: []appResponse{
			{Name: "test-app", RepoURL: "https://github.com/test-repo/gitops.git", Environments: []string{"dev"}},
		},
	}

	resp := applicationsToAppsResponse(apps, "https://github.com/test-repo/gitops")

	if diff := cmp.Diff(want, resp); diff != "" {
		t.Fatal(fmt.Errorf("WANT[%v] != RECEIVED[%v], diff=%s", want, resp, diff))
	}

	resp = applicationsToAppsResponse(apps, "https://github.com/test-repo/gitops.git")

	if diff := cmp.Diff(want, resp); diff != "" {
		t.Fatal(fmt.Errorf("WANT[%v] != RECEIVED[%v], diff=%s", want, resp, diff))
	}
}
