package service

import (
	"context"
	"testing"
)

func newRegisterService() (*RegisterService, *memUserRepo, *memSessionRepo, *memCodes) {
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	codes := newMemCodes()
	tokens := &stubTokens{}
	s := NewRegisterService(users, sessions, tokens, codes, plainPhone{}, plainHasher{}, &seqID{})
	return s, users, sessions, codes
}

func TestRegisterSuccess(t *testing.T) {
	ctx := context.Background()
	s, _, sessions, codes := newRegisterService()

	codes.Save(ctx, mustPhone(t, "13800000000"), "123456")

	res, err := s.Register(ctx, RegisterInput{
		Phone:      "13800000000",
		Code:       "123456",
		Password:   "secret",
		PublicName: "alice",
		DeviceID:   "dev-1",
	})
	if err != nil {
		t.Fatalf("Register err: %v", err)
	}
	if res.User.PublicName().String() != "alice" {
		t.Fatalf("public name = %q", res.User.PublicName())
	}
	if res.SessionID == "" {
		t.Fatal("session id empty")
	}
	if res.Tokens.AccessToken == "" {
		t.Fatal("access token empty")
	}
	if len(sessions.sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions.sessions))
	}
}

func TestRegisterInvalidCode(t *testing.T) {
	ctx := context.Background()
	s, _, _, codes := newRegisterService()

	codes.Save(ctx, mustPhone(t, "13800000000"), "123456")

	_, err := s.Register(ctx, RegisterInput{
		Phone:      "13800000000",
		Code:       "wrong",
		Password:   "secret",
		PublicName: "alice",
		DeviceID:   "dev-1",
	})
	if err != ErrInvalidCode {
		t.Fatalf("err = %v, want ErrInvalidCode", err)
	}
}

func TestRegisterCodeOneTime(t *testing.T) {
	ctx := context.Background()
	s, _, _, codes := newRegisterService()

	codes.Save(ctx, mustPhone(t, "13800000000"), "123456")

	if _, err := s.Register(ctx, RegisterInput{Phone: "13800000000", Code: "123456", Password: "secret", PublicName: "alice", DeviceID: "dev-1"}); err != nil {
		t.Fatalf("first register err: %v", err)
	}
	// 第二次用同一验证码必须失败（一次性，SC-002）
	if _, err := s.Register(ctx, RegisterInput{Phone: "13900000000", Code: "123456", Password: "secret", PublicName: "bob", DeviceID: "dev-2"}); err != ErrInvalidCode {
		t.Fatalf("second register err = %v, want ErrInvalidCode (code already used)", err)
	}
}

func TestRegisterDuplicatePhone(t *testing.T) {
	ctx := context.Background()
	s, _, _, codes := newRegisterService()

	codes.Save(ctx, mustPhone(t, "13800000000"), "123456")
	if _, err := s.Register(ctx, RegisterInput{Phone: "13800000000", Code: "123456", Password: "secret", PublicName: "alice", DeviceID: "dev-1"}); err != nil {
		t.Fatalf("first register err: %v", err)
	}

	codes.Save(ctx, mustPhone(t, "13800000000"), "654321")
	_, err := s.Register(ctx, RegisterInput{Phone: "13800000000", Code: "654321", Password: "secret2", PublicName: "alice2", DeviceID: "dev-2"})
	if err != ErrPhoneTaken {
		t.Fatalf("err = %v, want ErrPhoneTaken", err)
	}
}
