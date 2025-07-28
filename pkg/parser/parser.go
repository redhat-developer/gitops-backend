package parser

import (
	"github.com/go-git/go-git/v5"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resource"
	fs "sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/redhat-developer/gitops-backend/pkg/gitfs"
)

// ParseFromGit takes a go-git CloneOptions struct and a filepath, and extracts
// the service configuration from there.
func ParseFromGit(path string, opts *git.CloneOptions) ([]*Resource, error) {
	gfs, err := gitfs.NewInMemoryFromOptions(opts)
	if err != nil {
		return nil, err
	}
	return parseConfig(path, gfs)
}

func parseConfig(path string, files fs.FileSystem) ([]*Resource, error) {

	// Run performs a kustomization.
	// It reads given path from the given file system, interprets it as
	// a kustomization.yaml file, perform the kustomization it represents,
	// and return the resulting resources.
	kt := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	r, err := kt.Run(files, path)
	if err != nil {
		return nil, err
	}
	if len(r.Resources()) == 0 {
		return nil, nil
	}

	conv, err := newUnstructuredConverter()
	if err != nil {
		return nil, err
	}
	resources := []*Resource{}
	for _, v := range r.Resources() {
		resources = append(resources, extractResource(conv, v))
	}
	return resources, nil
}

// convert the Kustomize Resource into an internal representation, extracting
// the images if possible.
//
// If this is an unknown type (to the converter) no images will be extracted.
func extractResource(conv *unstructuredConverter, res *resource.Resource) *Resource {
	c := convert(res)
	g := c.GroupVersionKind()
	r := &Resource{
		Name:      c.GetName(),
		Namespace: c.GetNamespace(),
		Group:     g.Group,
		Version:   g.Version,
		Kind:      g.Kind,
		Labels:    c.GetLabels(),
	}
	t, err := conv.fromUnstructured(c)
	if err != nil {
		return r
	}
	r.Images = extractImages(t)
	return r
}

// convert converts a Kustomize resource into a generic Unstructured resource
// which the unstructured converter uses to create resources from.
func convert(r *resource.Resource) *unstructured.Unstructured {
	res, _ := r.Map()
	return &unstructured.Unstructured{
		Object: res,
	}
}
