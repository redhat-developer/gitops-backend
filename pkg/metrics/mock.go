package metrics

var _ Interface = (*MockMetrics)(nil)

// MockMetrics is a type that provides a simple counter for metrics for test
// purposes.
type MockMetrics struct {
	APICalls       int
	FailedAPICalls int
}

// NewMock creates and returns a MockMetrics.
func NewMock() *MockMetrics {
	return &MockMetrics{}
}

// CountAPICall records outgoing API calls to upstream services.
func (m *MockMetrics) CountAPICall(name string) {
	m.APICalls++
}

// CountFailedAPICall records failed outgoing API calls to upstream services.
func (m *MockMetrics) CountFailedAPICall(name string) {
	m.FailedAPICalls++
}
