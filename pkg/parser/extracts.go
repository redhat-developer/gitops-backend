package parser

import (
	ocpappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/gitops-backend/internal/sets"
)

// Deployments, DeploymentConfigs, StatefulSets, DaemonSets, Jobs, CronJobs
func extractImages(v interface{}) []string {
	switch k := v.(type) {
	case *appsv1.Deployment:
		return extractImagesFromPodTemplateSpec(k.Spec.Template)
	case *appsv1.StatefulSet:
		return extractImagesFromPodTemplateSpec(k.Spec.Template)
	case *appsv1.DaemonSet:
		return extractImagesFromPodTemplateSpec(k.Spec.Template)
	case *batchv1.Job:
		return extractImagesFromPodTemplateSpec(k.Spec.Template)
	case *batchv1beta1.CronJob:
		return extractImagesFromPodTemplateSpec(k.Spec.JobTemplate.Spec.Template)
	case *ocpappsv1.DeploymentConfig:
		return extractImagesFromPodTemplateSpec(*k.Spec.Template)
	}
	return nil
}

func extractImagesFromPodTemplateSpec(p corev1.PodTemplateSpec) []string {
	images := sets.NewStringSet()
	for _, c := range p.Spec.InitContainers {
		images.Add(c.Image)
	}
	for _, c := range p.Spec.Containers {
		images.Add(c.Image)
	}
	return images.Elements()
}
