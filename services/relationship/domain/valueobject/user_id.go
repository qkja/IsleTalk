package valueobject

import "errors"

// UserId 是不可变的内部账号 ID（承接 identity-svc 的 UUIDv7 字符串）。
type UserId struct {
	value string
}

var ErrEmptyUserId = errors.New("user id must not be empty")

func NewUserId(v string) (UserId, error) {
	if v == "" {
		return UserId{}, ErrEmptyUserId
	}
	return UserId{value: v}, nil
}

func (u UserId) String() string { return u.value }

func (u UserId) Equal(o UserId) bool { return u.value == o.value }
