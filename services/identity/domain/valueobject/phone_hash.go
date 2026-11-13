package valueobject

import "errors"

// PhoneHash 是手机号的加盐哈希（R-3）。明文手机号 MUST NOT 进入领域层；
// 领域层只见到该哈希值，用于注册去重与登录定位。
type PhoneHash struct {
	value string
}

var ErrEmptyPhoneHash = errors.New("phone hash must not be empty")

func NewPhoneHash(v string) (PhoneHash, error) {
	if v == "" {
		return PhoneHash{}, ErrEmptyPhoneHash
	}
	return PhoneHash{value: v}, nil
}

func (p PhoneHash) String() string { return p.value }
