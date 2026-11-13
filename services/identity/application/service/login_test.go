package service

import (
	"context"
	"testing"
)

func newLoginService() (*LoginService, *memUserRepo) {
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	tokens := &stubTokens{}
	s := NewLoginService(users, sessions, tokens, plainPhone{}, plainHasher{}, &seqID{})
	return s, users
}

func TestLoginSuccess(t *testing.T) {
	ctx := context.Background()
	s, users := newLoginService()

	// 预置一个已注册用户
	u, _ := newUserForTest(t, "13800000000", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	res, err := s.Login(ctx, LoginInput{Phone: "13800000000", Password: "secret", DeviceID: "dev-1"})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.User.PublicName().String() != "alice" {
		t.Fatalf("public name = %q", res.User.PublicName())
	}
	if res.Tokens.AccessToken == "" {
		t.Fatal("access token empty")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	ctx := context.Background()
	s, users := newLoginService()

	u, _ := newUserForTest(t, "13800000000", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	_, err := s.Login(ctx, LoginInput{Phone: "13800000000", Password: "wrong", DeviceID: "dev-1"})
	if err != ErrInvalidCredential {
		t.Fatalf("err = %v, want ErrInvalidCredential", err)
	}
}

func TestLoginUnregistered(t *testing.T) {
	ctx := context.Background()
	s, _ := newLoginService()

	// 未注册手机号：必须与密码错误返回相同错误（防枚举，SC-003）
	_, err := s.Login(ctx, LoginInput{Phone: "13900000000", Password: "whatever", DeviceID: "dev-1"})
	if err != ErrInvalidCredential {
		t.Fatalf("err = %v, want ErrInvalidCredential (anti-enumeration)", err)
	}
}
