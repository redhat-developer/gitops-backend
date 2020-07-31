package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestStringSet(t *testing.T) {
	s := stringSet{}

	s.add("testing1")
	s.add("testing1")
	s.add("testing2")

	got := s.elements()

	want := []string{"testing1", "testing2"}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("set failed:\n%s", diff)
	}
}

func TestExtractImagesFromDeployment(t *testing.T) {
	d := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Name: "redis", Image: "redis:6-alpine"},
						{Name: "redis-test", Image: "redis:6-alpine"},
					},
					Containers: []corev1.Container{
						{Name: "http", Image: "example/http-api"},
					},
				},
			},
		},
	}

	images := extractImagesFromDeployment(d)

	want := []string{"example/http-api", "redis:6-alpine"}
	if diff := cmp.Diff(want, images); diff != "" {
		t.Fatalf("set failed:\n%s", diff)
	}
}
