package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rhd-gitops-examples/gitops-backend/test"
)

func TestParseConfigFromURLWithNoFile(t *testing.T) {
	app, err := ParseConfigFromURL(".", "testdata")
	if app != nil {
		t.Errorf("did not expect to parse an app: %#v", app)
	}
	if !test.MatchError(t, "unable to find one of", err) {
		t.Fatalf("error didn't match: %s", err)
	}
}

func TestParseConfigFromURL(t *testing.T) {
	parseTests := []struct {
		filename    string
		description string
		want        *Config
	}{
		{
			"testdata/app1",
			"empty kustomization",
			nil,
		},
		{
			"testdata/go-demo",
			"completely local - paths refer to relative paths",
			&Config{
				Apps: []*App{
					{
						Name: "go-demo",
						Services: []*Service{
							{Name: "go-demo-http", Replicas: 1, Images: []string{"bigkevmcd/go-demo:876ecb3"}},
							{Name: "redis", Replicas: 1, Images: []string{"redis:6-alpine"}},
						},
					},
				},
			},
		},
		{
			"testdata/app2",
			"local file refers to a remote path - THIS COULD BREAK",
			&Config{
				Apps: []*App{
					{
						Name: "taxi",
						Services: []*Service{
							{Name: "taxi", Replicas: 1, Images: []string{"quay.io/kmcdermo/taxi:147036"}},
						},
					},
				},
			},
		},
	}

	for _, tt := range parseTests {
		app, err := ParseConfigFromURL(".", tt.filename)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(tt.want, app); diff != "" {
			t.Errorf("%s failed to parse:\n%s", tt.filename, diff)
		}
	}
}

func TestAppName(t *testing.T) {
	redis := map[string]string{"app.kubernetes.io/name": "redis", "app.kubernetes.io/part-of": "go-demo"}
	name := appName(redis)

	if name != "go-demo" {
		t.Fatalf("got %#v, want %#v", name, "go-demo")
	}
}

func TestExtractService(t *testing.T) {
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
		"spec": map[string]interface{}{
			"replicas": int64(1),
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "redis:6-alpine",
							"name":  "redis",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 6379,
								},
							},
						},
					},
				},
			},
		},
	}

	svc := extractService(redisMap)
	want := &Service{
		Name:      "redis",
		Namespace: "test-env",
		Replicas:  1,
		Images:    []string{"redis:6-alpine"},
	}
	assertCmp(t, want, svc, "failed to match service")
}

func assertCmp(t *testing.T, want, got interface{}, msg string) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf(msg+":\n%s", diff)
	}
}
