package parser

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// UnstructuredConverter handles conversions between unstructured.Unstructured and Contour types
type UnstructuredConverter struct {
	// scheme holds an initializer for converting Unstructured to a type
	scheme *runtime.Scheme
}

// NewUnstructuredConverter returns a new UnstructuredConverter initialized
func NewUnstructuredConverter() (*UnstructuredConverter, error) {
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		appsv1.AddToScheme,
	}

	uc := &UnstructuredConverter{
		scheme: runtime.NewScheme(),
	}

	if err := schemeBuilder.AddToScheme(uc.scheme); err != nil {
		return nil, err
	}

	return uc, nil
}

// FromUnstructured converts an unstructured.Unstructured to typed struct. If obj
// is not an unstructured.Unstructured it is returned without further processing.
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
