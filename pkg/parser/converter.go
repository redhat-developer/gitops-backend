package parser

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// UnstructuredConverter handles conversions between unstructured.Unstructured
// and some core Kubernetes resource types.
type UnstructuredConverter struct {
	scheme *runtime.Scheme
}

// NewUnstructuredConverter creates and returns a new UnstructuredConverter.
func NewUnstructuredConverter() (*UnstructuredConverter, error) {
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		appsv1.AddToScheme,
		batchv1.AddToScheme,
		batchv1beta1.AddToScheme,
	}

	uc := &UnstructuredConverter{
		scheme: runtime.NewScheme(),
	}

	if err := schemeBuilder.AddToScheme(uc.scheme); err != nil {
		return nil, err
	}
	return uc, nil
}

// FromUnstructured converts an unstructured.Unstructured to typed struct.
//
// If obj is not an unstructured.Unstructured it is returned without further processing.
// If unable to convert using the Kind of obj, then an error is returned.
func (c *UnstructuredConverter) FromUnstructured(obj interface{}) (interface{}, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return obj, nil
	}

	newObj, err := c.scheme.New(u.GetObjectKind().GroupVersionKind())
	if err != nil {
		return nil, err
	}
	return newObj, c.scheme.Convert(obj, newObj, nil)
}
