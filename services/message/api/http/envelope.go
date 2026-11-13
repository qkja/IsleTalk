package http

import (
	"errors"

	"isletalk/message/application/service"
)

// Meta 是信封的元信息，承载机器可读错误码（宪法 B 组）。
type Meta struct {
	Code string `json:"code,omitempty"`
}

// Envelope 是统一信封（宪法原则 III）：{success, data, message, meta}。
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message *string     `json:"message"`
	Meta    *Meta       `json:"meta,omitempty"`
}

func ok(data interface{}) Envelope {
	return Envelope{Success: true, Data: data}
}

func fail(code, msg string) Envelope {
	m := msg
	return Envelope{Success: false, Message: &m, Meta: &Meta{Code: code}}
}

func errorEnvelope(err error) Envelope {
	code, msg := errorCode(err)
	return fail(code, msg)
}

// errorCode 映射用例错误到稳定错误码（错误码注册表见 002）。
func errorCode(err error) (string, string) {
	switch {
	case errors.Is(err, service.ErrNotFriends):
		return "MSG_NOT_FRIENDS", "非好友，无法发送消息"
	case errors.Is(err, service.ErrBlocked):
		return "MSG_BLOCKED", "已被拉黑，无法发送消息"
	case errors.Is(err, service.ErrMessageNotFound):
		return "MSG_NOT_FOUND", "消息不存在"
	case errors.Is(err, service.ErrConversationNotFound):
		return "CONV_NOT_FOUND", "会话不存在"
	case errors.Is(err, service.ErrRecallWindowExpired):
		return "MSG_RECALL_WINDOW_EXPIRED", "撤回时间窗已过"
	case errors.Is(err, service.ErrNotSender):
		return "MSG_NOT_SENDER", "非消息发送者，无法撤回"
	case errors.Is(err, service.ErrInvalidMessageType):
		return "ARG_INVALID", "消息类型无效"
	case errors.Is(err, service.ErrNotParticipant):
		return "MSG_NOT_PARTICIPANT", "不是会话参与者"
	default:
		return "SYS_UNAVAILABLE", "服务暂不可用"
	}
}
