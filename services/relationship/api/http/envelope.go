package http

import (
	"errors"

	"isletalk/relationship/application/service"
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
	case errors.Is(err, service.ErrSelfRequest):
		return "ARG_INVALID", "不能添加自己"
	case errors.Is(err, service.ErrStrangerAddDisabled):
		return "RES_STRANGER_ADD_DISABLED", "对方未开启陌生人添加"
	case errors.Is(err, service.ErrRequestExists):
		return "RES_REQUEST_EXISTS", "好友申请已存在"
	case errors.Is(err, service.ErrRequestNotFound):
		return "RES_NOT_FOUND", "好友申请不存在"
	case errors.Is(err, service.ErrBlocked):
		return "RES_BLOCKED", "已被拉黑"
	case errors.Is(err, service.ErrAlreadyFriends):
		return "RES_ALREADY_FRIENDS", "已是好友"
	default:
		return "SYS_UNAVAILABLE", "服务暂不可用"
	}
}
