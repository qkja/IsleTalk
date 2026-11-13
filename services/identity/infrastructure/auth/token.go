package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
	"isletalk/identity/infrastructure/config"
)

// TokenIssuer 用 HMAC 签发的简单令牌（access ≤15min、refresh ≤7day）。
// 生产可替换为 JWT + JWKS；此处保持无第三方依赖、可编译。
type TokenIssuer struct {
	secret []byte
}

func NewTokenIssuer() TokenIssuer {
	secret := os.Getenv("IDENTITY_TOKEN_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret" // 仅本地开发占位
	}
	return TokenIssuer{secret: []byte(secret)}
}

type claims struct {
	UserID string `json:"sub"`
	Aud    string `json:"aud"`
	Exp    int64  `json:"exp"`
}

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

func (t TokenIssuer) Issue(_ context.Context, userID valueobject.UserId) (port.TokenPair, error) {
	accessTTL := config.Duration("IDENTITY_TOKEN_ACCESS_TTL", 15*time.Minute)
	refreshTTL := config.Duration("IDENTITY_TOKEN_REFRESH_TTL", 7*24*time.Hour)
	access, err := t.sign(userID, "access", accessTTL)
	if err != nil {
		return port.TokenPair{}, err
	}
	refresh, err := t.sign(userID, "refresh", refreshTTL)
	if err != nil {
		return port.TokenPair{}, err
	}
	return port.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (t TokenIssuer) sign(userID valueobject.UserId, aud string, ttl time.Duration) (string, error) {
	payload, err := json.Marshal(claims{
		UserID: userID.String(),
		Aud:    aud,
		Exp:    time.Now().Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	b64 := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(b64))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return b64 + "." + sig, nil
}

func (t TokenIssuer) Verify(_ context.Context, token string) (valueobject.UserId, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return valueobject.UserId{}, ErrTokenInvalid
	}
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if subtleEqual(parts[1], expected) != 1 {
		return valueobject.UserId{}, ErrTokenInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return valueobject.UserId{}, ErrTokenInvalid
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return valueobject.UserId{}, ErrTokenInvalid
	}
	if time.Now().Unix() > c.Exp {
		return valueobject.UserId{}, ErrTokenExpired
	}
	uid, err := valueobject.NewUserId(c.UserID)
	if err != nil {
		return valueobject.UserId{}, ErrTokenInvalid
	}
	return uid, nil
}

func (t TokenIssuer) Revoke(_ context.Context, _ valueobject.UserId) error {
	// 简单实现：撤销由会话仓储的 RevokeAllByUser 承载（令牌无状态）。
	return nil
}

func subtleEqual(a, b string) int {
	if len(a) != len(b) {
		return 0
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	if v == 0 {
		return 1
	}
	return 0
}
