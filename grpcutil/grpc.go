package grpcutil

import (
	"admission/customerrors"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func BuildSecurityDialOption() (grpc.DialOption, bool, error) {
	if os.Getenv("GRPC_TLS") == "true" {
		tlsCredentials, errTLS := LoadTLSCredentials()
		if errTLS != nil {
			return nil, false, customerrors.Wrap(http.StatusInternalServerError, errTLS, "loadTLSCredentials cannot read certificates")
		}
		return grpc.WithTransportCredentials(tlsCredentials), true, nil
	}

	// if security is not enabled, use the insecure version
	return grpc.WithTransportCredentials(insecure.NewCredentials()), false, nil
}

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(filepath.Join(os.Getenv("CERT_FOLDER_PATH"), "ca-cert.pem"))
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS13,
	}

	return credentials.NewTLS(config), nil
}
