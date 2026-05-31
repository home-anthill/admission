package customerrors

import "testing"

func TestGrpcSendErrorReturnsMessage(t *testing.T) {
	err := GrpcSendError{
		Status:  ConnectionError,
		Message: "cannot connect",
	}

	if got := err.Error(); got != "cannot connect" {
		t.Fatalf("Error() = %q, want %q", got, "cannot connect")
	}
}
