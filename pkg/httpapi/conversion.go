package httpapi

import (
	"sort"

	argoV1aplha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
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

func applicationsToAppsResponse(appSet []*argoV1aplha1.Application) *appsResponse {
	appsMap := make(map[string]appResponse)
	for _, app := range appSet {
		appName := app.ObjectMeta.Labels["app.kubernetes.io/name"]
		if appName == "" {
			appName = app.ObjectMeta.Name
		}
		if appResp, ok := appsMap[appName]; !ok {
			appsMap[appName] = appResponse{
				Name:         appName,
				RepoURL:      app.Spec.Source.RepoURL,
				Environments: []string{app.Spec.Destination.Namespace},
			}
		} else {
			appResp.Environments = append(appResp.Environments, app.Spec.Destination.Namespace)
			appsMap[appName] = appResp
		}
	}

	var apps []appResponse
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return &appsResponse{Apps: apps}
}
