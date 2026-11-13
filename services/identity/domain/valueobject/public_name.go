package valueobject

import "errors"

// PublicName 是公开账号名：全局唯一、可修改、可被搜索（FR-013/014）。
// 不变量：非空、经唯一性与保留词校验后才可构造为有效值。
type PublicName struct {
	value string
}

var (
	ErrEmptyPublicName   = errors.New("public name must not be empty")
	ErrPublicNameTooLong = errors.New("public name exceeds maximum length")
)

const MaxPublicNameLen = 32

func NewPublicName(v string) (PublicName, error) {
	if v == "" {
		return PublicName{}, ErrEmptyPublicName
	}
	if len([]rune(v)) > MaxPublicNameLen {
		return PublicName{}, ErrPublicNameTooLong
	}
	return PublicName{value: v}, nil
}

func (p PublicName) String() string { return p.value }
