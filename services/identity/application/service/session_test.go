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
	svc := NewDeactivateService(users, sessions, newMemVerificationRepo())

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

func TestVerificationRevokeClearsDigest(t *testing.T) {
	ctx := context.Background()
	repo := newMemVerificationRepo()
	svc := NewVerificationService(repo)

	uid := idVal(t, "u1")
	v, err := svc.Submit(ctx, uid, "idcard", "encrypted-digest", now(), 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Submit err: %v", err)
	}
	if v.IDDigest() == "" {
		t.Fatal("digest should be set before revoke")
	}

	if err := svc.Revoke(ctx, uid); err != nil {
		t.Fatalf("Revoke err: %v", err)
	}
	revoked, err := repo.FindByUser(ctx, uid)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if revoked.IDDigest() != "" {
		t.Fatalf("digest = %q, want empty after revoke (FR-024)", revoked.IDDigest())
	}
}

func TestDeactivateRevokesVerification(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	sessions := newMemSessionRepo()
	verifs := newMemVerificationRepo()
	svc := NewDeactivateService(users, sessions, verifs)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	verifs.Save(ctx, entity.NewVerification(u.ID(), "idcard", "digest", now(), now().Add(30*24*time.Hour)))

	if _, err := svc.Deactivate(ctx, u.ID()); err != nil {
		t.Fatalf("Deactivate err: %v", err)
	}

	// 注销后实名应被撤回（FR-025/028）
	v, err := verifs.FindByUser(ctx, u.ID())
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if v.IsSubmitted() {
		t.Fatal("verification should be revoked after deactivate")
	}
}

func TestSessionValidate(t *testing.T) {
	ctx := context.Background()
	sessions := newMemSessionRepo()
	svc := NewSessionService(sessions)

	sessions.Create(ctx, entity.NewSession("s1", idVal(t, "u1"), "d1"))

	got, err := svc.Validate(ctx, "s1")
	if err != nil {
		t.Fatalf("Validate err: %v", err)
	}
	if got.ID() != "s1" {
		t.Fatalf("id = %q, want s1", got.ID())
	}

	// 撤销后立即拒绝（FR-011）
	if err := svc.Revoke(ctx, "s1"); err != nil {
		t.Fatalf("Revoke err: %v", err)
	}
	if _, err := svc.Validate(ctx, "s1"); err != ErrSessionRevoked {
		t.Fatalf("err = %v, want ErrSessionRevoked", err)
	}
}

func TestSessionValidateNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewSessionService(newMemSessionRepo())
	if _, err := svc.Validate(ctx, "missing"); err != ErrSessionRevoked {
		t.Fatalf("err = %v, want ErrSessionRevoked", err)
	}
}

func TestPurgeExpiredVerifications(t *testing.T) {
	ctx := context.Background()
	repo := newMemVerificationRepo()
	svc := NewVerificationService(repo)

	now := time.Now()
	// 已过期（retention_until 在过去）
	repo.Save(ctx, entity.NewVerification(idVal(t, "u1"), "idcard", "digest1", now.Add(-48*time.Hour), now.Add(-24*time.Hour)))
	// 未过期
	repo.Save(ctx, entity.NewVerification(idVal(t, "u2"), "idcard", "digest2", now, now.Add(30*24*time.Hour)))

	n, err := svc.PurgeExpired(ctx, now)
	if err != nil {
		t.Fatalf("PurgeExpired err: %v", err)
	}
	if n != 1 {
		t.Fatalf("deleted = %d, want 1", n)
	}
	if _, err := repo.FindByUser(ctx, idVal(t, "u2")); err != nil {
		t.Fatalf("u2 should remain: %v", err)
	}
	if _, err := repo.FindByUser(ctx, idVal(t, "u1")); err == nil {
		t.Fatal("u1 should be purged")
	}
}
