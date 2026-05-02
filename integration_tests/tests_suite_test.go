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
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration tests")
}
