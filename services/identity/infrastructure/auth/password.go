package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"isletalk/identity/domain/valueobject"
)

// PasswordHasher 用 Argon2id 哈希/校验密码（R-1，宪法 E 组）。
// 参数 m=19456 (19 MiB) / t=2 / p=1，单次 <1s。
type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasher { return PasswordHasher{} }

const (
	argonTime    = 2
	argonMemory  = 19456
	argonThreads = 1
	argonKeyLen  = 32
	saltLen      = 16
)

// Hash 生成带参数与盐的 Argon2id 哈希（PHC 风格字符串）。
func (PasswordHasher) Hash(plain string) (valueobject.PasswordHash, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return valueobject.PasswordHash{}, fmt.Errorf("generate salt: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return valueobject.NewPasswordHash(encoded)
}

// Verify 校验明文与哈希是否匹配（常数时间比较）。
func (PasswordHasher) Verify(plain string, hash valueobject.PasswordHash) bool {
	parts := strings.Split(hash.String(), "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(plain), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
