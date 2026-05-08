package integration_tests_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	if err := os.Setenv("ENV", "testing"); err != nil {
		t.Fatalf("cannot force ENV=testing: %v", err)
	}
	if err := os.Setenv("API_TOKEN_HASH_SECRET", "integration-test-api-token-hash-secret"); err != nil {
		t.Fatalf("cannot set API_TOKEN_HASH_SECRET: %v", err)
	}
	if err := os.Setenv("API_TOKEN_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef"); err != nil {
		t.Fatalf("cannot set API_TOKEN_ENCRYPTION_KEY: %v", err)
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}
