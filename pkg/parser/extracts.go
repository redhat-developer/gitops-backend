package parser

import (
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/pkg/gvk"
)

func extractImages(conv *UnstructuredConverter, g gvk.Gvk, v *unstructured.Unstructured) []string {
	d, err := conv.FromUnstructured(v)
	if err != nil {
		return nil
	}
	// Deployments, DeploymentConfigs, StatefulSets, DaemonSets, Jobs, CronJobs
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
