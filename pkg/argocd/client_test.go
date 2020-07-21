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

// 2020/07/20 18:46:12 Payload = *models.V1alpha1Application
// 2020/07/20 18:46:12 Name = "guestbook"
// 2020/07/20 18:46:12        &{Message: Status:Healthy}
// 2020/07/20 18:46:12 Resources = *models.V1alpha1ResourceStatus &{Group:apps Health:0xc0004b5a80 Hook:false Kind:Deployment Name:guestbook-ui Namespace:default RequiresPruning:false Status:Synced Version:v1}
// 2020/07/20 18:46:12 Health = *models.V1alpha1HealthStatus &{Message: Status:Healthy}
// 2020/07/20 18:46:12 Resources = *models.V1alpha1ResourceStatus &{Group: Health:0xc0004b5aa0 Hook:false Kind:Service Name:guestbook-ui Namespace:default RequiresPruning:false Status:Synced Version:v1}
// 2020/07/20 18:46:12 Health = *models.V1alpha1HealthStatus &{Message: Status:Healthy}
