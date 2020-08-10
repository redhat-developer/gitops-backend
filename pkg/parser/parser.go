package parser

import (
	"github.com/go-git/go-git/v5"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/target"

	"github.com/rhd-gitops-example/gitops-backend/pkg/gitfs"
)

const (
	serviceLabel = "app.kubernetes.io/name"
	appLabel     = "app.kubernetes.io/part-of"
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
	k8sfactory := k8sdeps.NewFactory()
	ldr, err := loader.NewLoader(path, files)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = ldr.Cleanup()
		if err != nil {
			panic(err)
		}
	}()
	kt, err := target.NewKustTarget(ldr, k8sfactory.ResmapF, k8sfactory.TransformerF)
	if err != nil {
		return nil, err
	}
	r, err := kt.MakeCustomizedResMap()
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, nil
	}

	conv, err := newUnstructuredConverter()
	if err != nil {
		return nil, err
	}
	resources := []*Resource{}
	for k, v := range r {
		resources = append(resources, extractResource(conv, k.Gvk(), v))
	}
	return resources, nil
}

func extractResource(conv *unstructuredConverter, g gvk.Gvk, v *resource.Resource) *Resource {
	m := v.Map()
	meta := m["metadata"].(map[string]interface{})
	r := &Resource{
		Name:      mapString("name", meta),
		Namespace: mapString("namespace", meta),
		Group:     g.Group,
		Version:   g.Version,
		Kind:      g.Kind,
		Labels:    mapStringMap("labels", meta),
	}
	r.Images = extractImages(conv, convert(v))
	return r
}

// TODO: These should be extracted from the parsed resource.
func mapString(k string, v map[string]interface{}) string {
	s, ok := v[k].(string)
	if !ok {
		return ""
	}
	return s
}

func mapStringMap(key string, meta map[string]interface{}) map[string]string {
	s, ok := meta[key].(map[string]interface{})
	if !ok {
		return map[string]string{}
	}
	items := map[string]string{}
	for k, v := range s {
		items[k] = v.(string)
	}
	return items
}

// convert converts a Kustomize resource into a generic Unstructured resource
// which which the unstructured converter uses to create resources from.
func convert(r *resource.Resource) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: r.Map(),
	}
}
