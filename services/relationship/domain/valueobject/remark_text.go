package valueobject

import "errors"

// RemarkText 是好友备注文本（FR-005，单向，仅设置方可见）。
type RemarkText struct {
	value string
}

var (
	ErrEmptyRemark   = errors.New("remark must not be empty")
	ErrRemarkTooLong = errors.New("remark exceeds maximum length")
)

const MaxRemarkLen = 50

func NewRemarkText(v string) (RemarkText, error) {
	if v == "" {
		return RemarkText{}, ErrEmptyRemark
	}
	if len([]rune(v)) > MaxRemarkLen {
		return RemarkText{}, ErrRemarkTooLong
	}
	return RemarkText{value: v}, nil
}

func (r RemarkText) String() string { return r.value }
