package parser

import (
	"sort"

	ocpappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Deployments, DeploymentConfigs, StatefulSets, DaemonSets, Jobs, CronJobs
func extractImages(conv *unstructuredConverter, v *unstructured.Unstructured) []string {
	d, err := conv.fromUnstructured(v)
	if err != nil {
		return nil
	}
	switch k := d.(type) {
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
	images := stringSet{}
	for _, c := range p.Spec.InitContainers {
		images.add(c.Image)
	}
	for _, c := range p.Spec.Containers {
		images.add(c.Image)
	}
	return images.elements()
}

type stringSet map[string]bool

func (s stringSet) add(n string) {
	s[n] = true
}

func (s stringSet) elements() []string {
	e := []string{}
	for k := range s {
		e = append(e, k)
	}
	sort.Strings(e)
	return e
}
