package customerrors

import (
	"errors"
	"net/http"
	"testing"
)

func TestWrapKeepsCodeMessageAndCause(t *testing.T) {
	cause := errors.New("database down")

	err := Wrap(http.StatusServiceUnavailable, cause, "cannot register")

	var wrapped ErrorWrapper
	if !errors.As(err, &wrapped) {
		t.Fatal("Wrap() did not return an ErrorWrapper")
	}
	if wrapped.Code != http.StatusServiceUnavailable {
		t.Fatalf("Code = %d, want %d", wrapped.Code, http.StatusServiceUnavailable)
	}
	if wrapped.Message != "cannot register" {
		t.Fatalf("Message = %q, want %q", wrapped.Message, "cannot register")
	}
	if got := err.Error(); got != "database down" {
		t.Fatalf("Error() = %q, want %q", got, "database down")
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped error should unwrap to the original cause")
	}
}

func TestErrorWrapperUsesMessageWhenCauseIsMissing(t *testing.T) {
	err := ErrorWrapper{Message: "fallback message"}

	if got := err.Error(); got != "fallback message" {
		t.Fatalf("Error() = %q, want %q", got, "fallback message")
	}
	if err.Unwrap() != nil {
		t.Fatal("Unwrap() should be nil when no cause is present")
	}
}
