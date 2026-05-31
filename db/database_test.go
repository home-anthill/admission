package db

import "testing"

func TestGetDbNameUsesTestDatabaseWhenEnvIsTesting(t *testing.T) {
	t.Setenv("ENV", "testing")

	if got := getDbName(); got != "api-server-test" {
		t.Fatalf("getDbName() = %q, want %q", got, "api-server-test")
	}
}

func TestGetDbNameUsesDefaultDatabase(t *testing.T) {
	t.Setenv("ENV", "prod")

	if got := getDbName(); got != "api-server" {
		t.Fatalf("getDbName() = %q, want %q", got, "api-server")
	}
}
