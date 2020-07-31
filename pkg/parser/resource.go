package parser

// Object a basic object for a Kubernetes object.
type Resource struct {
	Group     string            `json:"group"`
	Version   string            `json:"version"`
	Kind      string            `json:"kind"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"-"`
	Images    []string          `json:"-"`
}
