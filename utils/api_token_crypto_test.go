package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestHashAPITokenIsDeterministic(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", strings.Repeat("s", 32))

	first, err := HashAPIToken("token")
	if err != nil {
		t.Fatalf("HashAPIToken() error = %v", err)
	}
	second, err := HashAPIToken("token")
	if err != nil {
		t.Fatalf("HashAPIToken() second error = %v", err)
	}
	if first != second {
		t.Fatalf("HashAPIToken() = %q then %q, want same value", first, second)
	}
}

func TestHashAPITokenRequiresSecret(t *testing.T) {
	t.Setenv("API_TOKEN_HASH_SECRET", "")

	hash, err := HashAPIToken("token")

	if err == nil {
		t.Fatal("HashAPIToken() error = nil, want error")
	}
	if hash != "" {
		t.Fatalf("HashAPIToken() hash = %q, want empty", hash)
	}
}

func TestEncryptAndDecryptAPITokenWithRawKey(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", strings.Repeat("k", 32))

	encrypted, err := EncryptAPIToken("secret-token")
	if err != nil {
		t.Fatalf("EncryptAPIToken() error = %v", err)
	}

	decrypted, err := DecryptAPIToken(encrypted)
	if err != nil {
		t.Fatalf("DecryptAPIToken() error = %v", err)
	}
	if decrypted != "secret-token" {
		t.Fatalf("DecryptAPIToken() = %q, want %q", decrypted, "secret-token")
	}
}

func TestEncryptAndDecryptAPITokenWithBase64Key(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", key)

	encrypted, err := EncryptAPIToken("secret-token")
	if err != nil {
		t.Fatalf("EncryptAPIToken() error = %v", err)
	}

	decrypted, err := DecryptAPIToken(encrypted)
	if err != nil {
		t.Fatalf("DecryptAPIToken() error = %v", err)
	}
	if decrypted != "secret-token" {
		t.Fatalf("DecryptAPIToken() = %q, want %q", decrypted, "secret-token")
	}
}

func TestEncryptAndDecryptAPITokenWithStandardBase64Key(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", key)

	encrypted, err := EncryptAPIToken("secret-token")
	if err != nil {
		t.Fatalf("EncryptAPIToken() error = %v", err)
	}

	decrypted, err := DecryptAPIToken(encrypted)
	if err != nil {
		t.Fatalf("DecryptAPIToken() error = %v", err)
	}
	if decrypted != "secret-token" {
		t.Fatalf("DecryptAPIToken() = %q, want %q", decrypted, "secret-token")
	}
}

func TestEncryptAPITokenRequiresKey(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "")

	encrypted, err := EncryptAPIToken("token")

	if err == nil {
		t.Fatal("EncryptAPIToken() error = nil, want error")
	}
	if encrypted != "" {
		t.Fatalf("EncryptAPIToken() = %q, want empty", encrypted)
	}
}

func TestEncryptAPITokenRequiresValidKey(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", "too-short")

	encrypted, err := EncryptAPIToken("token")

	if err == nil {
		t.Fatal("EncryptAPIToken() error = nil, want error")
	}
	if encrypted != "" {
		t.Fatalf("EncryptAPIToken() = %q, want empty", encrypted)
	}
}

func TestDecryptAPITokenRejectsTamperedCiphertext(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", strings.Repeat("k", 32))

	encrypted, err := EncryptAPIToken("secret-token")
	if err != nil {
		t.Fatalf("EncryptAPIToken() error = %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("cannot decode encrypted token: %v", err)
	}
	raw[len(raw)-1] ^= 1

	decrypted, err := DecryptAPIToken(base64.RawURLEncoding.EncodeToString(raw))

	if err == nil {
		t.Fatal("DecryptAPIToken() error = nil, want error")
	}
	if decrypted != "" {
		t.Fatalf("DecryptAPIToken() = %q, want empty", decrypted)
	}
}

func TestDecryptAPITokenRejectsInvalidInput(t *testing.T) {
	t.Setenv("API_TOKEN_ENCRYPTION_KEY", strings.Repeat("k", 32))

	tests := []struct {
		name      string
		encrypted string
	}{
		{name: "not base64", encrypted: "%"},
		{name: "too short", encrypted: base64.RawURLEncoding.EncodeToString([]byte("short"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := DecryptAPIToken(tt.encrypted)
			if err == nil {
				t.Fatal("DecryptAPIToken() error = nil, want error")
			}
			if decrypted != "" {
				t.Fatalf("DecryptAPIToken() = %q, want empty", decrypted)
			}
		})
	}
}
