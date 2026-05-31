package initialization

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestPrintEnvRequiresHashSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")

	err := printEnv(zap.NewNop().Sugar())

	if err == nil {
		t.Fatal("printEnv() error = nil, want error")
	}
}

func TestPrintEnvRequiresLongHashSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "short")

	err := printEnv(zap.NewNop().Sugar())

	if err == nil {
		t.Fatal("printEnv() error = nil, want error")
	}
}

func TestPrintEnvAcceptsValidHashSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", strings.Repeat("s", 32))

	err := printEnv(zap.NewNop().Sugar())

	if err != nil {
		t.Fatalf("printEnv() error = %v", err)
	}
}
