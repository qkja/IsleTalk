package persistence

import (
	"strings"
	"testing"
)

// setTestKey 注入 32 字节测试密钥（生产密钥经环境变量 IDENTITY_ID_ENC_KEY 注入，MUST NOT 硬编码）。
func setTestKey(t *testing.T) {
	t.Helper()
	t.Setenv("IDENTITY_ID_ENC_KEY", strings.Repeat("a", 32))
}

func TestIDCipherRoundTrip(t *testing.T) {
	setTestKey(t)
	c, err := NewIDCipher()
	if err != nil {
		t.Fatalf("NewIDCipher: %v", err)
	}

	plain := "110101199001011234"
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == plain || enc == "" {
		t.Fatal("ciphertext should differ from plaintext")
	}

	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != plain {
		t.Fatalf("roundtrip = %q, want %q", dec, plain)
	}
}

func TestIDCipherUniqueNonce(t *testing.T) {
	setTestKey(t)
	c, err := NewIDCipher()
	if err != nil {
		t.Fatalf("NewIDCipher: %v", err)
	}
	a, _ := c.Encrypt("same")
	b, _ := c.Encrypt("same")
	if a == b {
		t.Fatal("same plaintext should produce different ciphertext (unique nonce)")
	}
}

func TestIDCipherMissingKey(t *testing.T) {
	t.Setenv("IDENTITY_ID_ENC_KEY", "")
	if _, err := NewIDCipher(); err == nil {
		t.Fatal("missing key should error")
	}
}

func TestIDCipherInvalidKeyLength(t *testing.T) {
	t.Setenv("IDENTITY_ID_ENC_KEY", "too-short")
	if _, err := NewIDCipher(); err == nil {
		t.Fatal("32-byte key requirement should be enforced")
	}
}

func TestIDCipherDecryptInvalid(t *testing.T) {
	setTestKey(t)
	c, _ := NewIDCipher()
	if _, err := c.Decrypt("not-base64!!"); err == nil {
		t.Fatal("invalid ciphertext should error")
	}
}
