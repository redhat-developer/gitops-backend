package parser

import (
	"github.com/go-git/go-git/v5"

	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/target"

	"github.com/rhd-gitops-examples/gitops-backend/pkg/gitfs"
	"github.com/rhd-gitops-examples/gitops-backend/pkg/resource"
)

const (
	serviceLabel = "app.kubernetes.io/name"
	appLabel     = "app.kubernetes.io/part-of"
)

// ParseFromGit takes a go-git CloneOptions struct and a filepath, and extracts
// the service configuration from there.
func ParseFromGit(path string, opts *git.CloneOptions) ([]*resource.Resource, error) {
	gfs, err := gitfs.NewInMemoryFromOptions(opts)
	if err != nil {
		return nil, err
	}
	return parseConfig(path, gfs)
}

func parseConfig(path string, files fs.FileSystem) ([]*resource.Resource, error) {
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

	resources := []*resource.Resource{}
	for k, v := range r {
		resources = append(resources, extractResource(k.Gvk(), v.Map()))
	}
	return resources, nil
}

func extractResource(g gvk.Gvk, v map[string]interface{}) *resource.Resource {
	meta := v["metadata"].(map[string]interface{})
	return &resource.Resource{
		Name:      mapString("name", meta),
		Namespace: mapString("namespace", meta),
		Group:     g.Group,
		Version:   g.Version,
		Kind:      g.Kind,
	}
}

func mapString(k string, v map[string]interface{}) string {
	s, ok := v[k].(string)
	if !ok {
		return ""
	}
	return s
}
