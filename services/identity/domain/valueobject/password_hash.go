package valueobject

import "errors"

// PasswordHash 是密码的 Argon2id 哈希（带参数与盐，R-1）。
// 领域层不比较明文密码，只经由 PasswordVerifier port 校验。
type PasswordHash struct {
	value string
}

var ErrEmptyPasswordHash = errors.New("password hash must not be empty")

func NewPasswordHash(v string) (PasswordHash, error) {
	if v == "" {
		return PasswordHash{}, ErrEmptyPasswordHash
	}
	return PasswordHash{value: v}, nil
}

func (p PasswordHash) String() string { return p.value }
