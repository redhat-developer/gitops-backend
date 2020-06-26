package httpapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestWithNoAuthorizationHeader(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc(""))
	req := makeTokenRequest("")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	assertHTTPError(t, resp, http.StatusForbidden, "Authentication required")
}

func TestRequestWithBadHeader(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc(""))
	req := makeTokenRequest("Bearer ")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()

	assertHTTPError(t, resp, http.StatusForbidden, "Authentication required")
}

func TestRequestWithBadPrefix(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc(""))
	req := makeTokenRequest("Authentication token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()

	assertHTTPError(t, resp, http.StatusForbidden, "Authentication required")
}

func TestRequestWithAuthorizationHeader(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc("testing-token"))
	req := makeTokenRequest("Bearer testing-token")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestWithAuthorizationHeaderSetsTokenInContext(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc("testing-token"))
	req := makeTokenRequest("Bearer testing-token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("incorrect status code, got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if s := strings.TrimSpace(string(body)); s != "testing-token" {
		t.Fatalf("got %s, want %s", s, "testing-token\n")
	}

}

func makeTokenRequest(token string) *http.Request {
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	if token != "" {
		req.Header.Set(authHeader, token)
	}
	return req
}

func makeTestFunc(wantedToken string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := AuthToken(r.Context())
		if !ok {
			token = "success"
		}
		fmt.Fprintln(w, token)
	})
}

func assertHTTPError(t *testing.T, resp *http.Response, code int, want string) {
	t.Helper()
	if resp.StatusCode != code {
		t.Errorf("status code didn't match, got %d, want %d", resp.StatusCode, code)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if s := strings.TrimSpace(string(b)); s != want {
		t.Fatalf("got %s, want %s", s, want)
	}
}
