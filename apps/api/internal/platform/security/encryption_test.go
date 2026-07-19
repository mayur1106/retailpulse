package security

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptorRoundTrip(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32)))
	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}
	ciphertext, err := encryptor.Encrypt([]byte("amazon-refresh-token"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	plaintext, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plaintext) != "amazon-refresh-token" {
		t.Fatalf("unexpected plaintext: %s", plaintext)
	}
}
