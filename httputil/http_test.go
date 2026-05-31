package httputil

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetReturnsStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("not ready"))
	}))
	defer server.Close()

	statusCode, body, err := Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if statusCode != http.StatusTeapot {
		t.Fatalf("Get() statusCode = %d, want %d", statusCode, http.StatusTeapot)
	}
	if body != "not ready" {
		t.Fatalf("Get() body = %q, want %q", body, "not ready")
	}
}

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

func TestGetReturnsWrappedErrorWhenRequestFails(t *testing.T) {
	statusCode, body, err := Get("://bad-url")
	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
	if statusCode != -1 {
		t.Fatalf("Get() statusCode = %d, want -1", statusCode)
	}
	if body != "" {
		t.Fatalf("Get() body = %q, want empty", body)
	}
}

func TestGetReturnsWrappedErrorWhenBodyReadFails(t *testing.T) {
	restoreHTTPClient(t, &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       failingBody{},
			}, nil
		}),
	})

	statusCode, body, err := Get("http://example.test")

	if err == nil {
		t.Fatal("Get() error = nil, want error")
	}
	if statusCode != http.StatusOK {
		t.Fatalf("Get() statusCode = %d, want %d", statusCode, http.StatusOK)
	}
	if body != "" {
		t.Fatalf("Get() body = %q, want empty", body)
	}
}

func TestPostReturnsStatusAndBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	defer server.Close()

	statusCode, body, err := Post(server.URL, []byte(`{"ok":true}`))
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if statusCode != http.StatusCreated {
		t.Fatalf("Post() statusCode = %d, want %d", statusCode, http.StatusCreated)
	}
	if body != "created" {
		t.Fatalf("Post() body = %q, want %q", body, "created")
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

func TestPostReturnsWrappedErrorWhenRequestFails(t *testing.T) {
	statusCode, body, err := Post("://bad-url", []byte(`{}`))
	if err == nil {
		t.Fatal("Post() error = nil, want error")
	}
	if statusCode != -1 {
		t.Fatalf("Post() statusCode = %d, want -1", statusCode)
	}
	if body != "" {
		t.Fatalf("Post() body = %q, want empty", body)
	}
}

func TestPostReturnsWrappedErrorWhenBodyReadFails(t *testing.T) {
	restoreHTTPClient(t, &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       failingBody{},
			}, nil
		}),
	})

	statusCode, body, err := Post("http://example.test", []byte(`{}`))

	if err == nil {
		t.Fatal("Post() error = nil, want error")
	}
	if statusCode != http.StatusCreated {
		t.Fatalf("Post() statusCode = %d, want %d", statusCode, http.StatusCreated)
	}
	if body != "" {
		t.Fatalf("Post() body = %q, want empty", body)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (failingBody) Close() error {
	return nil
}

var _ io.ReadCloser = failingBody{}

func restoreHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()

	original := httpClient
	httpClient = client
	t.Cleanup(func() {
		httpClient = original
	})
}
