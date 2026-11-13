package service

import (
	"context"
	"testing"
)

func newLoginService() (*LoginService, *memUserRepo) {
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	tokens := &stubTokens{}
	s := NewLoginService(users, sessions, tokens, newMemCodes(), plainPhone{}, plainHasher{}, &seqID{})
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

func TestLoginByCodeSuccess(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	codes := newMemCodes()
	s := NewLoginService(users, newMemSessionRepo(), &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	u, _ := newUserForTest(t, "13800000000", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	codes.Save(ctx, mustPhone(t, "13800000000"), "123456")

	res, err := s.LoginByCode(ctx, LoginByCodeInput{Phone: "13800000000", Code: "123456", DeviceID: "dev-1"})
	if err != nil {
		t.Fatalf("LoginByCode err: %v", err)
	}
	if res.User.PublicName().String() != "alice" {
		t.Fatalf("public name = %q", res.User.PublicName())
	}
	if res.Tokens.AccessToken == "" {
		t.Fatal("access token empty")
	}
}

func TestLoginByCodeInvalidCode(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	s := NewLoginService(users, newMemSessionRepo(), &stubTokens{}, newMemCodes(), plainPhone{}, plainHasher{}, &seqID{})

	u, _ := newUserForTest(t, "13800000000", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	_, err := s.LoginByCode(ctx, LoginByCodeInput{Phone: "13800000000", Code: "wrong", DeviceID: "dev-1"})
	if err != ErrInvalidCredential {
		t.Fatalf("err = %v, want ErrInvalidCredential", err)
	}
}

func TestLoginByCodeUnregistered(t *testing.T) {
	ctx := context.Background()
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "13900000000"), "123456")
	s := NewLoginService(newMemUserRepo(), newMemSessionRepo(), &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	// 未注册手机号：与验证码错误统一返回 ErrInvalidCredential（防枚举）
	_, err := s.LoginByCode(ctx, LoginByCodeInput{Phone: "13900000000", Code: "123456", DeviceID: "dev-1"})
	if err != ErrInvalidCredential {
		t.Fatalf("err = %v, want ErrInvalidCredential (anti-enumeration)", err)
	}
}
