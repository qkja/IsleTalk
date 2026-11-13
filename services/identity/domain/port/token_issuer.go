package port

import (
	"context"

	"isletalk/identity/domain/valueobject"
)

// TokenIssuer 签发/校验/撤销访问令牌与刷新令牌（R-2）。
// 领域层只依赖此接口，具体用 JWT + JWKS 由 infrastructure 实现。
type TokenIssuer interface {
	// Issue 为指定用户签发访问令牌（access ≤15min）与刷新令牌（refresh ≤7day）。
	Issue(ctx context.Context, userID valueobject.UserId) (TokenPair, error)
	// Verify 校验访问令牌，返回其中的用户 ID；无效返回错误。
	Verify(ctx context.Context, accessToken string) (valueobject.UserId, error)
	// Revoke 撤销刷新令牌（改密/注销全失效时调用）。
	Revoke(ctx context.Context, userID valueobject.UserId) error
}

// TokenPair 是一对访问/刷新令牌。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// VerificationCodeSender 下发验证码（R-5）。本地 mock，生产接云短信。
type VerificationCodeSender interface {
	// Send 向指定手机号哈希对应的手机号下发验证码。
	Send(ctx context.Context, phone valueobject.PhoneHash, code string) error
}
