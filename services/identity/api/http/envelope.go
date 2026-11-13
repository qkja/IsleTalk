package http

import (
	"errors"

	"isletalk/identity/application/service"
	"isletalk/identity/infrastructure/auth"
)

// Meta 是信封的元信息，承载机器可读错误码（宪法 B 组）。
type Meta struct {
	Code string `json:"code,omitempty"`
}

// Envelope 是统一信封（宪法原则 III）：{success, data, message, meta}。
// message 面向用户（安全文案），code 面向机器（唯一注册表）。
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message *string     `json:"message"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ok 构造成功信封。
func ok(data interface{}) Envelope {
	return Envelope{Success: true, Data: data}
}

// fail 构造失败信封（业务失败 MUST success=false + 稳定错误码）。
func fail(code, msg string) Envelope {
	m := msg
	return Envelope{Success: false, Message: &m, Meta: &Meta{Code: code}}
}

// errorEnvelope 把应用层错误映射为统一信封（错误码来自唯一注册表）。
// message 是面向用户的安全文案（FR-019 双语，经 i18n 按 locale 取），MUST NOT 含堆栈/内部细节。
func errorEnvelope(err error, locale string) Envelope {
	code, key := errorCode(err)
	return fail(code, auth.Message(key, locale))
}

// errorCode 映射应用层错误到稳定错误码 + i18n 文案 key（错误码注册表，见 002 contracts/error-registry.md）。
func errorCode(err error) (string, string) {
	switch {
	case errors.Is(err, service.ErrInvalidCode):
		return "RATE_VERIFY_CODE_COOLDOWN", "error_invalid_code"
	case errors.Is(err, service.ErrInvalidCredential):
		return "AUTH_UNAUTHENTICATED", "error_invalid_credential"
	case errors.Is(err, service.ErrPhoneTaken):
		return "RES_CONFLICT", "error_phone_taken"
	case errors.Is(err, service.ErrNameUnavailable):
		return "RES_CONFLICT", "error_name_unavailable"
	case errors.Is(err, service.ErrResetFailed):
		return "ARG_INVALID", "error_reset_failed"
	case errors.Is(err, service.ErrSessionRevoked):
		return "AUTH_TOKEN_EXPIRED", "error_session_revoked"
	default:
		return "SYS_UNAVAILABLE", "error_sys_unavailable"
	}
}
