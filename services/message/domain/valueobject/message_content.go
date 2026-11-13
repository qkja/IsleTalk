package valueobject

import "errors"

// MessageContent 是不可变的消息内容：文本正文或附件对象 key（FR-016/017/018）。
// 附件正文存对象存储，消息只存引用 key（plan Assumptions：客户端直传对象存储）。
type MessageContent struct {
	value string
}

// MaxContentBytes 是消息内容（文本正文或附件 key）的最大长度（宪法原则 I 输入边界校验）。
const MaxContentBytes = 8192

var (
	ErrEmptyContent   = errors.New("message content must not be empty")
	ErrContentTooLong = errors.New("message content exceeds limit")
)

func NewMessageContent(v string) (MessageContent, error) {
	if v == "" {
		return MessageContent{}, ErrEmptyContent
	}
	if len(v) > MaxContentBytes {
		return MessageContent{}, ErrContentTooLong
	}
	return MessageContent{value: v}, nil
}

func (c MessageContent) String() string { return c.value }
