package httpapi

import (
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	argoV1aplha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
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

func applicationsToAppsResponse(appSet []*argoV1aplha1.Application, repoURL string) *appsResponse {
	appsMap := make(map[string]appResponse)
	var appName string
	repoURL = strings.TrimSuffix(repoURL, ".git")

	for _, app := range appSet {
		appName = "" // Reset at beginning of the loop so no duplicates are added
		if repoURL != strings.TrimSuffix(app.Spec.Source.RepoURL, ".git") {
			log.Printf("repoURL[%v], does not match with Source Repo URL[%v]", repoURL, strings.TrimSuffix(app.Spec.Source.RepoURL, ".git"))
			continue
		}
		if app.ObjectMeta.Labels != nil {
			appName = app.ObjectMeta.Labels["app.kubernetes.io/name"]
		}

		if appName == "" {
			log.Println("AppName is empty")
			continue
		}

		var lastDeployed metav1.Time
		size := len(app.Status.History)
		if size > 0 {
			lastDeployed = app.Status.History[size-1].DeployedAt
		}
		var lastDeployedTime = lastDeployed.String()
		if lastDeployed.IsZero() {
			lastDeployedTime = ""
		}
		if appResp, ok := appsMap[appName]; !ok {
			appsMap[appName] = appResponse{
				Name:         appName,
				RepoURL:      app.Spec.Source.RepoURL,
				Environments: []string{app.Spec.Destination.Namespace},
				SyncStatus:   []string{string(app.Status.Sync.Status)},
				LastDeployed: []string{lastDeployedTime},
			}
		} else {
			appResp.Environments = append(appResp.Environments, app.Spec.Destination.Namespace)
			appResp.SyncStatus = append(appResp.SyncStatus, string(app.Status.Sync.Status))
			appResp.LastDeployed = append(appResp.LastDeployed, lastDeployed.String())
			appsMap[appName] = appResp
		}
	}

	var apps []appResponse
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return &appsResponse{Apps: apps}
}
