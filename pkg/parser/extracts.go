package parser

// eployments, deploymentconfigs, statefulsets, daemonsets, jobs, cronjobs
import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/pkg/gvk"
)

func extractImages(conv *UnstructuredConverter, g gvk.Gvk, v *unstructured.Unstructured) []string {
	d, err := conv.FromUnstructured(v)
	if err != nil {
		return nil
	}

	switch k := d.(type) {
	case *appsv1.Deployment:
		return extractImagesFromDeployment(k)
	}
	return nil
}

func extractImagesFromDeployment(d *appsv1.Deployment) []string {
	return extractImagesFromPodTemplateSpec(d.Spec.Template)
}

func extractImagesFromPodTemplateSpec(p corev1.PodTemplateSpec) []string {
	images := stringSet{}
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
	return e
}
