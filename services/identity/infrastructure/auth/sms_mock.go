package auth

import (
	"context"
	"fmt"

	"isletalk/identity/domain/valueobject"
)

// SMSVerificationCodeSender 是验证码下发适配器（R-5）。
// 本地开发 mock：不真发短信，仅打印；生产接云短信服务商。
type SMSVerificationCodeSender struct{}

func NewSMSVerificationCodeSender() SMSVerificationCodeSender { return SMSVerificationCodeSender{} }

func (SMSVerificationCodeSender) Send(_ context.Context, _ valueobject.PhoneHash, code string) error {
	// 本地 mock：不记录手机号明文，仅输出验证码供联调。
	fmt.Printf("[sms-mock] verification code: %s\n", code)
	return nil
}
