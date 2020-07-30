package parser

import (
	"github.com/go-git/go-git/v5"

	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/target"

	"github.com/rhd-gitops-example/gitops-backend/pkg/gitfs"
	"github.com/rhd-gitops-example/gitops-backend/pkg/resource"
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
	r := &resource.Resource{
		Name:      mapString("name", meta),
		Namespace: mapString("namespace", meta),
		Group:     g.Group,
		Version:   g.Version,
		Kind:      g.Kind,
		Labels:    mapStringMap("labels", meta),
	}
	if g.Kind == "Deployment" {
		r.Images = extractImagesFromDeployment(v)
	}
	return r
}

func extractImagesFromDeployment(v map[string]interface{}) []string {
	images := []string{}
	spec, ok := v["spec"].(map[string]interface{})
	if !ok {
		return images
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return images
	}
	templateSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return images
	}
	containers, ok := templateSpec["containers"].([]interface{})
	if !ok {
		return images
	}

	for _, v := range containers {
		container, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		images = append(images, mapString("image", container))
	}
	return images
}

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
