package metrics

// Interface implementations provide metrics for the system.
type Interface interface {
	// CountAPICall records API calls to the upstream hosting service.
	CountAPICall(name string)

	// CountFailedAPICall records failed API calls to the upstream hosting service.
	CountFailedAPICall(name string)
}
