package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

var _ Interface = (*PrometheusMetrics)(nil)

func TestCountAPICall(t *testing.T) {
	m := New("dsl", prometheus.NewRegistry())
	m.CountAPICall("file_contents")

	err := testutil.CollectAndCompare(m.apiCalls, strings.NewReader(`
# HELP dsl_api_calls_total Count of API Calls made
# TYPE dsl_api_calls_total counter
dsl_api_calls_total{kind="file_contents"} 1
`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCountFailedAPICall(t *testing.T) {
	m := New("dsl", prometheus.NewRegistry())
	m.CountFailedAPICall("commit_status")

	err := testutil.CollectAndCompare(m.failedAPICalls, strings.NewReader(`
# HELP dsl_failed_api_calls_total Count of failed API Calls made
# TYPE dsl_failed_api_calls_total counter
dsl_failed_api_calls_total{kind="commit_status"} 1
`))
	if err != nil {
		t.Fatal(err)
	}
}
