package service

import (
	"context"
	"testing"
	"time"

	"isletalk/identity/domain/entity"
)

func TestSessionRevokeAllByUser(t *testing.T) {
	ctx := context.Background()
	sessions := newMemSessionRepo()
	svc := NewSessionService(sessions)

	sess1 := entity.NewSession("s1", idVal(t, "u1"), "d1")
	sess2 := entity.NewSession("s2", idVal(t, "u1"), "d2")
	sess3 := entity.NewSession("s3", idVal(t, "u2"), "d3")
	sessions.Create(ctx, sess1)
	sessions.Create(ctx, sess2)
	sessions.Create(ctx, sess3)

	if err := svc.RevokeAllByUser(ctx, idVal(t, "u1")); err != nil {
		t.Fatalf("RevokeAllByUser err: %v", err)
	}

	got, _ := sessions.ListByUser(ctx, idVal(t, "u1"))
	for _, s := range got {
		if s.IsActive() {
			t.Fatalf("session %s should be revoked", s.ID())
		}
	}
	// 其他用户不受影响
	other, _ := sessions.ListByUser(ctx, idVal(t, "u2"))
	if len(other) != 1 || !other[0].IsActive() {
		t.Fatal("other user's session should remain active")
	}
}

func TestDeactivateRevokesSessions(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	svc := NewDeactivateService(users, sessions)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	sessions.Create(ctx, entity.NewSession("s1", u.ID(), "d1"))

	deactivated, err := svc.Deactivate(ctx, u.ID())
	if err != nil {
		t.Fatalf("Deactivate err: %v", err)
	}
	if deactivated.IsActive() {
		t.Fatal("user should be deactivated")
	}

	got, _ := sessions.ListByUser(ctx, u.ID())
	for _, s := range got {
		if s.IsActive() {
			t.Fatalf("session %s should be revoked after deactivate", s.ID())
		}
	}
}

func TestResetPasswordRevokesSessions(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	codes := newMemCodes()
	svc := NewResetPasswordService(users, sessions, codes, plainPhone{}, plainHasher{})

	u, _ := newUserForTest(t, "138", "old-secret", "alice")
	users.users[u.PhoneHash().String()] = u
	sessions.Create(ctx, entity.NewSession("s1", u.ID(), "d1"))
	codes.Save(ctx, mustPhone(t, "138"), "123456")

	if err := svc.Reset(ctx, "138", "123456", "new-secret"); err != nil {
		t.Fatalf("Reset err: %v", err)
	}

	// 密码已更新
	updated, _ := users.FindByPhoneHash(ctx, mustPhone(t, "138"))
	p := plainHasher{}
	if !p.Verify("new-secret", updated.PasswordHash()) {
		t.Fatal("password should be reset to new-secret")
	}
	// 旧会话全失效
	got, _ := sessions.ListByUser(ctx, u.ID())
	for _, s := range got {
		if s.IsActive() {
			t.Fatalf("session %s should be revoked after reset", s.ID())
		}
	}
}

func TestVerificationRevoke(t *testing.T) {
	ctx := context.Background()
	svc := NewVerificationService(newMemVerificationRepo())

	uid := idVal(t, "u1")
	v, err := svc.Submit(ctx, uid, "idcard", "encrypted-digest", now(), 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Submit err: %v", err)
	}
	if !v.IsSubmitted() {
		t.Fatal("verification should be submitted")
	}

	if err := svc.Revoke(ctx, uid); err != nil {
		t.Fatalf("Revoke err: %v", err)
	}
}
