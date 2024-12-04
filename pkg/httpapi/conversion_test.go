package httpapi

import (
	"fmt"
	"testing"
	"time"

	argoV1aplha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

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
				SyncStatus:   nil,
				LastDeployed: nil,
			},
		},
	}
	if diff := cmp.Diff(want, apps); diff != "" {
		t.Fatalf("failed to parse:\n%s", diff)
	}
}

func TestApplicationsToAppsResponse(t *testing.T) {
	var apps []*argoV1aplha1.Application
	app, _ := testArgoApplication("testdata/application.yaml")
	apps = append(apps, app)

	want := &appsResponse{
		Apps: []appResponse{
			{
				Name:         "test-app",
				RepoURL:      "https://github.com/test-repo/gitops.git",
				Environments: []string{"dev"},
				SyncStatus:   []string{"Synced"},
				LastDeployed: []string{time.Date(2021, time.Month(5), 15, 2, 12, 13, 0, time.UTC).Local().String()},
			},
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

func TestApplicationsToAppsResponseForTwoApps(t *testing.T) {
	var apps []*argoV1aplha1.Application
	app, _ := testArgoApplication("testdata/application.yaml")
	apps = append(apps, app)
	app2, _ := testArgoApplication("testdata/application2.yaml")
	apps = append(apps, app2)

	want := &appsResponse{
		Apps: []appResponse{
			{
				Name:         "test-app",
				RepoURL:      "https://github.com/test-repo/gitops.git",
				Environments: []string{"dev", "production"},
				SyncStatus:   []string{"Synced", "OutOfSync"},
				LastDeployed: []string{time.Date(2021, time.Month(5), 15, 2, 12, 13, 0, time.UTC).Local().String(),
					time.Date(2021, time.Month(5), 16, 1, 10, 35, 0, time.UTC).Local().String()},
			},
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
