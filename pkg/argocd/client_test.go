package argocd

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	apiclient "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client"
	appsvc "github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/client/application_service"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/argocd/models"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

func TestApplicationResources(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockClientService(ctrl)
	a := NewFromClient(&apiclient.Argocd{ApplicationService: m})
	params := appsvc.NewGetMixin8Params().WithName("testing")
	payload := &models.V1alpha1Application{
		Status: &models.V1alpha1ApplicationStatus{
			Resources: []*models.V1alpha1ResourceStatus{
				{Group: "apps", Kind: "Deployment", Name: "test-ui", Namespace: "default", Version: "v1"},
				{Group: "", Kind: "Service", Name: "test-ui", Namespace: "default", Version: "v1"},
			},
		},
	}
	m.
		EXPECT().
		GetMixin8(gomock.Eq(params)).
		Return(&appsvc.GetMixin8OK{Payload: payload}, nil)

	res, err := a.ApplicationResources("testing")

	if err != nil {
		t.Fatal(err)
	}
	want := []*resource.Resource{
		{Group: "apps", Version: "v1", Kind: "Deployment", Name: "test-ui", Namespace: "default"},
		{Group: "", Version: "v1", Kind: "Service", Name: "test-ui", Namespace: "default"},
	}
	if diff := cmp.Diff(want, res); diff != "" {
		t.Fatalf("ApplicationResources() failed:\n%s", diff)
	}
}
