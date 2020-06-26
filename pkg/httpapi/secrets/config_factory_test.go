package secrets

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/rest"
)

var _ RESTConfigFactory = (*K8sRESTConfigFactory)(nil)

func TestCreateClientFromToken(t *testing.T) {
	f := NewRESTConfigFactory(&rest.Config{}, false)

	cfg, err := f.Create("test-token")
	assertNoError(t, err)

	want := &rest.Config{BearerToken: "test-token"}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func TestCreateInsecureClientFromToken(t *testing.T) {
	f := NewRESTConfigFactory(&rest.Config{}, true)

	cfg, err := f.Create("new-token")
	assertNoError(t, err)

	want := &rest.Config{
		BearerToken: "new-token",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Fatalf("incorrect client config:\n%s", diff)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
