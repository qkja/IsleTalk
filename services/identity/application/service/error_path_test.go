package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errBoom = errors.New("boom")

func TestRegisterFindByPhoneFails(t *testing.T) {
	ctx := context.Background()
	users := failingUserRepo{err: errBoom}
	sessions := newMemSessionRepo()
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewRegisterService(users, sessions, &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Register(ctx, RegisterInput{Phone: "138", Code: "123456", Password: "s", PublicName: "a", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestLoginFindByPhoneFails(t *testing.T) {
	ctx := context.Background()
	users := failingUserRepo{err: errBoom}
	s := NewLoginService(users, newMemSessionRepo(), &stubTokens{}, newMemCodes(), plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Login(ctx, LoginInput{Phone: "138", Password: "s", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestDeactivateFindByIDFails(t *testing.T) {
	ctx := context.Background()
	users := failingUserRepo{err: errBoom}
	s := NewDeactivateService(users, newMemSessionRepo(), newMemVerificationRepo())

	_, err := s.Deactivate(ctx, idVal(t, "u-1"))
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestResetPasswordVerifyFails(t *testing.T) {
	ctx := context.Background()
	users := failingUserRepo{err: errBoom}
	s := NewResetPasswordService(users, newMemSessionRepo(), newMemCodes(), plainPhone{}, plainHasher{})

	err := s.Reset(ctx, "138", "code", "new")
	if err == nil {
		t.Fatal("Reset should fail when code not found")
	}
}

func TestRegisterCreateUserFails(t *testing.T) {
	ctx := context.Background()
	// FindByPhoneHash 返回 NotFound（正常），Create 返回错误
	users := failingCreateRepo{err: errBoom}
	sessions := newMemSessionRepo()
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewRegisterService(users, sessions, &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Register(ctx, RegisterInput{Phone: "138", Code: "123456", Password: "s", PublicName: "a", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestResetPasswordFindByPhoneFails(t *testing.T) {
	ctx := context.Background()
	users := failingUserRepo{err: errBoom}
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewResetPasswordService(users, newMemSessionRepo(), codes, plainPhone{}, plainHasher{})

	err := s.Reset(ctx, "138", "123456", "new")
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestRegisterCreateSessionFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	sessions := failingSessionRepo{err: errBoom}
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewRegisterService(users, sessions, &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Register(ctx, RegisterInput{Phone: "138", Code: "123456", Password: "s", PublicName: "a", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestRegisterIssueTokenFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	tokens := failingTokens{err: errBoom}
	s := NewRegisterService(users, sessions, tokens, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Register(ctx, RegisterInput{Phone: "138", Code: "123456", Password: "s", PublicName: "a", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestDeactivateRevokeSessionsFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	sessions := failingSessionRepo{err: errBoom}
	s := NewDeactivateService(users, sessions, newMemVerificationRepo())

	_, err := s.Deactivate(ctx, u.ID())
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestSessionRevokeFails(t *testing.T) {
	ctx := context.Background()
	svc := NewSessionService(failingSessionRepo{err: errBoom})
	if err := svc.Revoke(ctx, "s1"); err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestSessionRevokeAllByUserFails(t *testing.T) {
	ctx := context.Background()
	svc := NewSessionService(failingSessionRepo{err: errBoom})
	if err := svc.RevokeAllByUser(ctx, idVal(t, "u1")); err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestLoginCreateSessionFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	s := NewLoginService(users, failingSessionRepo{err: errBoom}, &stubTokens{}, newMemCodes(), plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Login(ctx, LoginInput{Phone: "138", Password: "secret", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestLoginIssueTokenFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	s := NewLoginService(users, newMemSessionRepo(), failingTokens{err: errBoom}, newMemCodes(), plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.Login(ctx, LoginInput{Phone: "138", Password: "secret", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestLoginByCodeCreateSessionFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewLoginService(users, failingSessionRepo{err: errBoom}, &stubTokens{}, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.LoginByCode(ctx, LoginByCodeInput{Phone: "138", Code: "123456", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestLoginByCodeIssueTokenFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	codes := newMemCodes()
	codes.Save(ctx, mustPhone(t, "138"), "123456")
	s := NewLoginService(users, newMemSessionRepo(), failingTokens{err: errBoom}, codes, plainPhone{}, plainHasher{}, &seqID{})

	_, err := s.LoginByCode(ctx, LoginByCodeInput{Phone: "138", Code: "123456", DeviceID: "d"})
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestDeactivateRevokeVerificationFails(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	s := NewDeactivateService(users, newMemSessionRepo(), failingVerificationRepo{err: errBoom})

	_, err := s.Deactivate(ctx, u.ID())
	if err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestSessionValidateFindFails(t *testing.T) {
	ctx := context.Background()
	svc := NewSessionService(failingSessionRepo{err: errBoom})
	if _, err := svc.Validate(ctx, "s1"); err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}

func TestPurgeExpiredFails(t *testing.T) {
	ctx := context.Background()
	svc := NewVerificationService(failingVerificationRepo{err: errBoom})
	if _, err := svc.PurgeExpired(ctx, time.Now()); err != errBoom {
		t.Fatalf("err = %v, want errBoom", err)
	}
}
