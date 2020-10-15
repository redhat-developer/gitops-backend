package httpapi

import (
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/pipelines.yaml")
	if err != nil {
		t.Fatal(err)
	}

	parsed := &config{}
	err = yaml.Unmarshal(b, parsed)
	if err != nil {
		t.Fatal(err)
	}

	want := &config{
		GitOpsURL: "https://example.com/demo/gitops.git",
		Environments: []*environment{
			{
				Name:    "dev",
				Cluster: "https://dev03.example.com",
				Apps: []*application{
					{
						Name: "taxi",
						Services: []service{
							{
								Name:      "gitops-demo",
								SourceURL: "https://example.com/demo/gitops-demo.git",
							},
						},
					},
				},
			},
			{
				Name:    "stage",
				Cluster: "https://kubernetes.default.svc",
			},
		},
	}

	if diff := cmp.Diff(want, parsed); diff != "" {
		t.Fatalf("parsed:\n%s", diff)
	}
}
