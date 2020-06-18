package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics is a wrapper around Prometheus metrics for counting
// events in the system.
type PrometheusMetrics struct {
	hooks          *prometheus.CounterVec
	invalidHooks   prometheus.Counter
	apiCalls       *prometheus.CounterVec
	failedAPICalls *prometheus.CounterVec
}

// New creates and returns a PrometheusMetrics initialised with prometheus
// counters.
func New(ns string, reg prometheus.Registerer) *PrometheusMetrics {
	pm := &PrometheusMetrics{}
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	pm.apiCalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ns,
		Name:      "api_calls_total",
		Help:      "Count of API Calls made",
	}, []string{"kind"})

	pm.failedAPICalls = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ns,
		Name:      "failed_api_calls_total",
		Help:      "Count of failed API Calls made",
	}, []string{"kind"})

	reg.MustRegister(pm.apiCalls)
	reg.MustRegister(pm.failedAPICalls)
	return pm
}

// CountAPICall records outgoing API calls to upstream services.
func (m *PrometheusMetrics) CountAPICall(name string) {
	m.apiCalls.With(prometheus.Labels{"kind": name}).Inc()
}

// CountFailedAPICall records failled outgoing API calls to upstream services.
func (m *PrometheusMetrics) CountFailedAPICall(name string) {
	m.failedAPICalls.With(prometheus.Labels{"kind": name}).Inc()
}
