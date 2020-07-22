package resource

// Resource is the basic metadata for a Kubernetes resource.
type Resource struct {
	Group        string `json:"group"`
	Version      string `json:"version"`
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	HealthStatus string `json:"healthStatus"`
}
