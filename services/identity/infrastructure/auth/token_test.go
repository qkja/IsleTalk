package auth

import (
	"context"
	"testing"
	"time"

	"isletalk/identity/domain/valueobject"
)

func TestTokenIssueVerify(t *testing.T) {
	issuer := NewTokenIssuer()
	uid, _ := valueobject.NewUserId("user-123")

	pair, err := issuer.Issue(context.Background(), uid)
	if err != nil {
		t.Fatalf("Issue err: %v", err)
	}

	got, err := issuer.Verify(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("Verify err: %v", err)
	}
	if got != uid {
		t.Fatalf("Verify got %v, want %v", got, uid)
	}
}

func TestTokenTamper(t *testing.T) {
	issuer := NewTokenIssuer()
	uid, _ := valueobject.NewUserId("user-123")
	pair, _ := issuer.Issue(context.Background(), uid)

	if _, err := issuer.Verify(context.Background(), pair.AccessToken+"x"); err != ErrTokenInvalid {
		t.Fatalf("tampered token err = %v, want ErrTokenInvalid", err)
	}
}

func TestTokenExpiry(t *testing.T) {
	issuer := NewTokenIssuer()
	uid, _ := valueobject.NewUserId("user-123")

	// 直接构造一个已过期的令牌（同包可访问私有方法）
	expired, err := issuer.sign(uid, "access", -time.Minute)
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	if _, err := issuer.Verify(context.Background(), expired); err != ErrTokenExpired {
		t.Fatalf("expired token err = %v, want ErrTokenExpired", err)
	}
}
