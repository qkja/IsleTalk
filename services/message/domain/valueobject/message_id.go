package valueobject

import "errors"

// MessageId 是不可变的消息幂等键（FR-007，客户端生成，重发去重）。
type MessageId struct {
	value string
}

var ErrEmptyMessageId = errors.New("message id must not be empty")

func NewMessageId(v string) (MessageId, error) {
	if v == "" {
		return MessageId{}, ErrEmptyMessageId
	}
	return MessageId{value: v}, nil
}

func (m MessageId) String() string { return m.value }

func (m MessageId) Equal(o MessageId) bool { return m.value == o.value }
