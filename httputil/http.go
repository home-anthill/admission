package httputil

import (
	"admission/customerrors"
	"bytes"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

const maxResponseBodyBytes = 64 * 1024

// Get performs an HTTP GET request with a timeout.
func Get(url string) (int, string, error) {
	response, err := httpClient.Get(url)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP GET API of the remote service")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes))
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot read HTTP GET response body")
	}
	return response.StatusCode, string(body), nil
}

// Post performs an HTTP POST request with a timeout.
func Post(url string, payloadJSON []byte) (int, string, error) {
	payloadBody := bytes.NewBuffer(payloadJSON)
	response, err := httpClient.Post(url, "application/json", payloadBody)
	if err != nil {
		return -1, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot call HTTP POST API of the remote service")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes))
	if err != nil {
		return response.StatusCode, "", customerrors.Wrap(http.StatusInternalServerError, err, "Cannot read HTTP POST response body")
	}
	return response.StatusCode, string(body), nil
}
