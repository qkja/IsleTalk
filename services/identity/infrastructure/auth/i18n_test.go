package auth

import (
	"testing"

	"isletalk/identity/domain/valueobject"
)

func TestI18nMessage(t *testing.T) {
	if got := Message("verify_code_sms", "zh-CN"); got == "" || got == "verify_code_sms" {
		t.Fatalf("zh-CN message = %q", got)
	}
	if got := Message("verify_code_sms", "en-US"); got == "" || got == "verify_code_sms" {
		t.Fatalf("en-US message = %q", got)
	}
	// 未知 locale 回退 zh-CN
	if Message("verify_code_sms", "fr") != Message("verify_code_sms", "zh-CN") {
		t.Fatal("unknown locale should fall back to zh-CN")
	}
	// 未知 key 返回 key 本身
	if Message("nonexistent", "zh-CN") != "nonexistent" {
		t.Fatal("unknown key should return itself")
	}
}

func TestSMSVerificationCodeSenderSend(t *testing.T) {
	s := NewSMSVerificationCodeSender()
	if err := s.Send(nil, mustPhoneHash(t), "123456"); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func mustPhoneHash(t *testing.T) valueobject.PhoneHash {
	t.Helper()
	h, err := valueobject.NewPhoneHash("phone:test")
	if err != nil {
		t.Fatal(err)
	}
	return h
}
