package httputil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCapsResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", maxResponseBodyBytes*2)))
	}))
	defer server.Close()

	statusCode, body, err := Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("Get() statusCode = %d, want %d", statusCode, http.StatusOK)
	}
	if len(body) != maxResponseBodyBytes {
		t.Fatalf("Get() body length = %d, want %d", len(body), maxResponseBodyBytes)
	}
}

func TestPostCapsResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", maxResponseBodyBytes*2)))
	}))
	defer server.Close()

	statusCode, body, err := Post(server.URL, []byte(`{}`))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("Post() statusCode = %d, want %d", statusCode, http.StatusOK)
	}
	if len(body) != maxResponseBodyBytes {
		t.Fatalf("Post() body length = %d, want %d", len(body), maxResponseBodyBytes)
	}
}
