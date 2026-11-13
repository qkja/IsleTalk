package persistence

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// IDCipher 用 AES-256-GCM 加密身份信息摘要（FR-022，宪法 E 组）。
// 密钥经环境变量注入，MUST NOT 硬编码；nonce 每次随机，MUST NOT 复用。
type IDCipher struct {
	aead cipher.AEAD
}

func NewIDCipher() (*IDCipher, error) {
	key := os.Getenv("IDENTITY_ID_ENC_KEY")
	if key == "" {
		key = "dev-insecure-32-byte-key-0000000000" // 仅本地开发占位（32 字节）
	}
	if len(key) != 32 {
		return nil, errors.New("IDENTITY_ID_ENC_KEY must be 32 bytes")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &IDCipher{aead: aead}, nil
}

// Encrypt 加密明文身份摘要，返回 base64(nonce || ciphertext)。
func (c *IDCipher) Encrypt(plain string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plain), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

// Decrypt 解密，仅在授权场景使用。
func (c *IDCipher) Decrypt(encoded string) (string, error) {
	data, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	ns := c.aead.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}
	plain, err := c.aead.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
