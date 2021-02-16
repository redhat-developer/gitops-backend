package applications

import (
	"context"
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	appclientset "github.com/argoproj/argo-cd/pkg/client/clientset/versioned"
	fakeappclientset "github.com/argoproj/argo-cd/pkg/client/clientset/versioned/fake"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var _ ApplicationGetter = (*KubeApplicationGetter)(nil)

var testID = types.NamespacedName{Name: "test-secret", Namespace: "test-ns"}

func TestGetApplication(t *testing.T) {
	g := createGetter(createApplication(testID))

	app, err := g.GetApplication(context.TODO(), "auth token", testID)
	if err != nil {
		t.Fatal(err)
	}

	want := createApplication(testID)
	assertApplicationEquals(t, app, want)
}

func TestApplicationWithMissingApplication(t *testing.T) {
	g := createGetter(createApplication(testID))
	unknownID := types.NamespacedName{Name: "unknown-app", Namespace: "test-ns"}

	_, err := g.GetApplication(context.TODO(), "auth token", unknownID)

	if err.Error() != `error getting application test-ns/unknown-app: applications.argoproj.io "unknown-app" not found` {
		t.Fatal(err)
	}
}

func TestListApplications(t *testing.T) {
	g := createGetter(createApplication(testID))

	apps, err := g.ListApplications(context.TODO(), "auth token", testID.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	want := []appv1.Application{*createApplication(testID)}
	if diff := cmp.Diff(want, apps, ignoreOptions()...); diff != "" {
		t.Fatalf("failed to get application:\n%s", diff)
	}
}

func TestListApplicationsRestrictsToNamespace(t *testing.T) {
	otherID := types.NamespacedName{Name: "other-app", Namespace: "other-ns"}
	g := createGetter(createApplication(testID), createApplication(otherID))

	apps, err := g.ListApplications(context.TODO(), "auth token", testID.Namespace)
	if err != nil {
		t.Fatal(err)
	}

	want := []appv1.Application{*createApplication(testID)}
	if diff := cmp.Diff(want, apps, ignoreOptions()...); diff != "" {
		t.Fatalf("failed to get application:\n%s", diff)
	}
}

func createApplication(id types.NamespacedName) *appv1.Application {
	return &appv1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id.Name,
			Namespace: id.Namespace,
		},
	}
}

func createGetter(o ...runtime.Object) *KubeApplicationGetter {
	g := New(&stubConfigFactory{})
	g.clientFactory = func(c *rest.Config) (appclientset.Interface, error) {
		if c.BearerToken == "auth token" {
			return fakeappclientset.NewSimpleClientset(o...), nil
		}
		return nil, errors.New("failed")
	}
	return g
}

type stubConfigFactory struct {
}

func (s *stubConfigFactory) Create(token string) (*rest.Config, error) {
	return &rest.Config{BearerToken: token}, nil
}

func assertApplicationEquals(t *testing.T, got, want *appv1.Application) {
	if diff := cmp.Diff(want, got, ignoreOptions()...); diff != "" {
		t.Fatalf("failed to get application:\n%s", diff)
	}
}

func ignoreOptions() []cmp.Option {
	return []cmp.Option{
		cmpopts.IgnoreUnexported(appv1.ApplicationDestination{}),
	}
}
