package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"isletalk/identity/domain/valueobject"
)

// PhoneHasher 用 HMAC-SHA256 将手机号加盐哈希（R-3）。
// 盐经环境变量注入，MUST NOT 硬编码；明文手机号随即弃。
type PhoneHasher struct {
	salt []byte
}

func NewPhoneHasher() PhoneHasher {
	salt := os.Getenv("IDENTITY_PHONE_SALT")
	if salt == "" {
		salt = "dev-insecure-salt" // 仅本地开发占位；生产 MUST 注入
	}
	return PhoneHasher{salt: []byte(salt)}
}

func (p PhoneHasher) Hash(phone string) valueobject.PhoneHash {
	mac := hmac.New(sha256.New, p.salt)
	mac.Write([]byte(phone))
	h, _ := valueobject.NewPhoneHash(hex.EncodeToString(mac.Sum(nil)))
	return h
}
