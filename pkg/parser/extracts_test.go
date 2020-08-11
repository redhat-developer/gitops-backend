package parser

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
)

func TestExtractImagesFromPodTemplateSpec(t *testing.T) {
	spec := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{Name: "redis", Image: "redis:6-alpine"},
				{Name: "redis-test", Image: "redis:6-alpine"},
			},
			Containers: []corev1.Container{
				{Name: "http", Image: "example/http-api"},
			},
		},
	}

	images := extractImagesFromPodTemplateSpec(spec)

	want := []string{"example/http-api", "redis:6-alpine"}
	if diff := cmp.Diff(want, images); diff != "" {
		t.Fatalf("set failed:\n%s", diff)
	}
}
