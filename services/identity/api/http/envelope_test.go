package http

import (
	"errors"
	"testing"

	"isletalk/identity/application/service"
)

func TestErrorCodeDefault(t *testing.T) {
	code, _ := errorCode(errors.New("some unknown error"))
	if code != "SYS_UNAVAILABLE" {
		t.Fatalf("default code = %q, want SYS_UNAVAILABLE", code)
	}
}

func TestErrorCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{service.ErrInvalidCode, "RATE_VERIFY_CODE_COOLDOWN"},
		{service.ErrInvalidCredential, "AUTH_UNAUTHENTICATED"},
		{service.ErrPhoneTaken, "RES_CONFLICT"},
		{service.ErrNameUnavailable, "RES_CONFLICT"},
		{service.ErrResetFailed, "ARG_INVALID"},
		{service.ErrSessionRevoked, "AUTH_TOKEN_EXPIRED"},
	}
	for _, c := range cases {
		code, _ := errorCode(c.err)
		if code != c.want {
			t.Fatalf("errorCode(%v) = %q, want %q", c.err, code, c.want)
		}
	}
}

func TestErrorEnvelopeI18n(t *testing.T) {
	en := errorEnvelope(service.ErrInvalidCredential, "en-US")
	if en.Message == nil || *en.Message != "Incorrect account or password" {
		t.Fatalf("en message = %v", en.Message)
	}
	zh := errorEnvelope(service.ErrInvalidCredential, "zh-CN")
	if zh.Message == nil || *zh.Message != "账号或密码错误" {
		t.Fatalf("zh message = %v", zh.Message)
	}
	// 未知 locale 回退 zh-CN
	fallback := errorEnvelope(service.ErrInvalidCredential, "fr-FR")
	if fallback.Message == nil || *fallback.Message != "账号或密码错误" {
		t.Fatalf("fallback message = %v", fallback.Message)
	}
}
