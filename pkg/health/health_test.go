package health

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPipelines(t *testing.T) {
	ts := makeServer(t)

	client := ts.Client()
	res, err := client.Get(ts.URL)

	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	want := fmt.Sprintf("{\"version\":\"%s\"}\n", GitRevision)
	if string(body) != want {
		t.Fatalf("failed to get the expected version, got %#v, want %#v", string(body), want)
	}
}

func makeServer(t *testing.T) *httptest.Server {
	ts := httptest.NewTLSServer(http.HandlerFunc(Handler))
	t.Cleanup(ts.Close)
	return ts
}
