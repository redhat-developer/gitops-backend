package httpapi

import (
	"sort"
	"strings"

	appv1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
)

// TODO: this should really import the config from the upstream and use it to
// unmarshal.
func pipelinesToAppsResponse(cfg *config) *appsResponse {
	appSet := map[string][]string{}
	for _, env := range cfg.Environments {
		for _, app := range env.Apps {
			envs, ok := appSet[app.Name]
			if !ok {
				envs = []string{}
			}
			envs = append(envs, env.Name)
			appSet[app.Name] = envs
		}
	}

	apps := []appResponse{}
	for k, v := range appSet {
		sort.Strings(v)
		apps = append(apps, appResponse{Name: k, RepoURL: cfg.GitOpsURL, Environments: v})
	}
	return &appsResponse{Apps: apps}
}

// TODO:
//  1. ensure that the name matches the naming convention "destination
//  NS-app-name"
//  2. What to do if the fields (Destination, Source) are unpopulated?
func appsToResponse(a []appv1.Application) *appsResponse {
	namesToEnvs := map[string][]string{}
	namesToURLs := map[string]string{}
	for _, v := range a {
		name := nameFromCR(v)
		envs, ok := namesToEnvs[name]
		if !ok {
			envs = []string{}
		}
		envs = append(envs, v.Spec.Destination.Namespace)
		namesToEnvs[name] = envs
		// This means that the app has to have the same gitops-repo URL
		// otherwise last one wins.
		namesToURLs[name] = repoURLFromCR(v)
	}

	apps := []appResponse{}
	for n, u := range namesToURLs {
		apps = append(apps, appResponse{
			Name:         n,
			RepoURL:      u,
			Environments: namesToEnvs[n],
		})
	}
	return &appsResponse{Apps: apps}
}

func repoURLFromCR(a appv1.Application) string {
	return a.Spec.Source.RepoURL
}

// This defines the naming convention that is parsed, we assume that the name of
// the apps is <environment>-app-<name>.
//
// This trims off the leading <environment>-app- part and returns it.
func nameFromCR(a appv1.Application) string {
	return strings.TrimPrefix(a.ObjectMeta.Name, a.Spec.Destination.Namespace+"-app-")
}
