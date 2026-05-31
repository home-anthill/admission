package grpcutil

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"admission/customerrors"
)

func TestBuildSecurityDialOptionUsesInsecureCredentialsByDefault(t *testing.T) {
	t.Setenv("GRPC_TLS", "false")

	option, secure, err := BuildSecurityDialOption()

	if err != nil {
		t.Fatalf("BuildSecurityDialOption() error = %v", err)
	}
	if secure {
		t.Fatal("BuildSecurityDialOption() secure = true, want false")
	}
	if option == nil {
		t.Fatal("BuildSecurityDialOption() option = nil")
	}
}

func TestBuildSecurityDialOptionWrapsTLSLoadError(t *testing.T) {
	t.Setenv("GRPC_TLS", "true")
	t.Setenv("CERT_FOLDER_PATH", t.TempDir())

	option, secure, err := BuildSecurityDialOption()

	if err == nil {
		t.Fatal("BuildSecurityDialOption() error = nil, want error")
	}
	if option != nil {
		t.Fatal("BuildSecurityDialOption() option should be nil on error")
	}
	if secure {
		t.Fatal("BuildSecurityDialOption() secure = true, want false on error")
	}

	var wrapped customerrors.ErrorWrapper
	if !errors.As(err, &wrapped) {
		t.Fatalf("BuildSecurityDialOption() error type = %T, want ErrorWrapper", err)
	}
	if wrapped.Code != http.StatusInternalServerError {
		t.Fatalf("wrapped.Code = %d, want %d", wrapped.Code, http.StatusInternalServerError)
	}
}

func TestLoadTLSCredentialsRejectsInvalidCA(t *testing.T) {
	certDir := t.TempDir()
	t.Setenv("CERT_FOLDER_PATH", certDir)
	if err := os.WriteFile(filepath.Join(certDir, "ca-cert.pem"), []byte("not a pem cert"), 0o600); err != nil {
		t.Fatalf("cannot write test certificate: %v", err)
	}

	credentials, err := LoadTLSCredentials()

	if err == nil {
		t.Fatal("LoadTLSCredentials() error = nil, want error")
	}
	if credentials != nil {
		t.Fatal("LoadTLSCredentials() credentials should be nil on error")
	}
}
