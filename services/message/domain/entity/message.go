package entity

import (
	"time"

	"isletalk/message/domain/valueobject"
)

// MessageType 是消息类型（FR-016/017/018；首段仅 text/image/file）。
type MessageType string

const (
	MessageText  MessageType = "text"
	MessageImage MessageType = "image"
	MessageFile  MessageType = "file"
)

// MessageStatus 是发送状态（FR-019，状态机：sent → delivered → read；failed 为发送失败终态）。
type MessageStatus string

const (
	MessageSent      MessageStatus = "sent"
	MessageDelivered MessageStatus = "delivered"
	MessageRead      MessageStatus = "read"
	MessageFailed    MessageStatus = "failed"
)

// Message 是消息聚合根。字段私有，状态迁移经显式方法（FR-019 状态机、FR-020 撤回）。
type Message struct {
	messageID      valueobject.MessageId
	conversationID string
	senderID       valueobject.UserId
	messageType    MessageType
	content        valueobject.MessageContent
	seq            uint64
	sentAt         time.Time
	status         MessageStatus
	recalled       bool
}

// NewMessage 创建新消息，初始状态 sent；服务端成功接收后由用例标记 delivered（FR-005）。
func NewMessage(
	messageID valueobject.MessageId,
	conversationID string,
	senderID valueobject.UserId,
	messageType MessageType,
	content valueobject.MessageContent,
	seq uint64,
) Message {
	return Message{
		messageID:      messageID,
		conversationID: conversationID,
		senderID:       senderID,
		messageType:    messageType,
		content:        content,
		seq:            seq,
		sentAt:         time.Now().UTC(),
		status:         MessageSent,
		recalled:       false,
	}
}

// ReconstructMessage 从持久化状态重建消息（仅基础设施仓储使用，勿用于新消息创建）。
func ReconstructMessage(
	messageID valueobject.MessageId,
	conversationID string,
	senderID valueobject.UserId,
	messageType MessageType,
	content valueobject.MessageContent,
	seq uint64,
	sentAt time.Time,
	status MessageStatus,
	recalled bool,
) Message {
	return Message{
		messageID:      messageID,
		conversationID: conversationID,
		senderID:       senderID,
		messageType:    messageType,
		content:        content,
		seq:            seq,
		sentAt:         sentAt,
		status:         status,
		recalled:       recalled,
	}
}

func (m Message) MessageID() valueobject.MessageId    { return m.messageID }
func (m Message) ConversationID() string              { return m.conversationID }
func (m Message) SenderID() valueobject.UserId        { return m.senderID }
func (m Message) Type() MessageType                   { return m.messageType }
func (m Message) Content() valueobject.MessageContent { return m.content }
func (m Message) Seq() uint64                         { return m.seq }
func (m Message) SentAt() time.Time                   { return m.sentAt }
func (m Message) Status() MessageStatus               { return m.status }
func (m Message) Recalled() bool                      { return m.recalled }

// IsSentBy 判断消息是否由指定用户发送（FR-020 撤回校验发送者身份）。
func (m Message) IsSentBy(userID valueobject.UserId) bool { return m.senderID.Equal(userID) }

// MarkDelivered 将状态从 sent 迁移到 delivered（服务端成功接收，FR-005）。
func (m *Message) MarkDelivered() {
	if m.status == MessageSent {
		m.status = MessageDelivered
	}
}

// MarkRead 将状态从 delivered 迁移到 read（接收方上报已读，FR-019）。
func (m *Message) MarkRead() {
	if m.status == MessageDelivered {
		m.status = MessageRead
	}
}

// MarkFailed 将状态迁移到 failed（发送失败终态，FR-019）。
func (m *Message) MarkFailed() {
	m.status = MessageFailed
}

// Recall 标记消息已撤回（FR-020；时间窗由服务端时钟判定，见应用层）。
func (m *Message) Recall() { m.recalled = true }
