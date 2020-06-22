package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestWithNoAuthorizationHeader(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc(""))
	req := makeRequest("")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "Authentication required\n" {
		t.Fatalf("got %s, want %s", body, "Authentication required\n")
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusForbidden)
	}
}

func TestRequestWithBadHeader(t *testing.T) {
	t.Skip()
}

func TestRequestWithAuthorizationHeader(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc("testing-token"))
	req := makeRequest("Bearer testing-token")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %v, want %v", resp.StatusCode, http.StatusOK)
	}
}

func TestRequestWithAuthorizationHeaderSetsTokenInContext(t *testing.T) {
	handler := AuthenticationMiddleware(makeTestFunc("testing-token"))
	req := makeRequest("Bearer testing-token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "testing-token\n" {
		t.Fatalf("got %s, want %s", body, "testing-token\n")
	}
}

func makeRequest(token string) *http.Request {
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
