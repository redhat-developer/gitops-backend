package parser

import (
	ocpappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// unstructuredConverter handles conversions between unstructured.Unstructured
// and some core Kubernetes resource types.
type unstructuredConverter struct {
	scheme *runtime.Scheme
}

// newUnstructuredConverter creates and returns a new UnstructuredConverter.
func newUnstructuredConverter() (*unstructuredConverter, error) {
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		appsv1.AddToScheme,
		batchv1.AddToScheme,
		batchv1beta1.AddToScheme,
		ocpappsv1.AddToScheme,
	}

	uc := &unstructuredConverter{
		scheme: runtime.NewScheme(),
	}

	if err := schemeBuilder.AddToScheme(uc.scheme); err != nil {
		return nil, err
	}
	return uc, nil
}

// fromUnstructured converts an unstructured.Unstructured to typed struct.
//
// If unable to convert using the Kind of obj, then an error is returned.
func (c *unstructuredConverter) fromUnstructured(o *unstructured.Unstructured) (interface{}, error) {
	newObj, err := c.scheme.New(o.GetObjectKind().GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newObj, c.scheme.Convert(o, newObj, nil)
}
