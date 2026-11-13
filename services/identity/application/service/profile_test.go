package service

import (
	"context"
	"testing"

	"isletalk/identity/domain/entity"
)

func TestProfileUpdate(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	svc := NewProfileService(users)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	got, err := svc.UpdateProfile(ctx, u.ID(), "Alice", "http://a/1.png", "hello")
	if err != nil {
		t.Fatalf("UpdateProfile err: %v", err)
	}
	if got.Nickname() != "Alice" || got.AvatarURL() != "http://a/1.png" || got.Signature() != "hello" {
		t.Fatalf("profile not updated: %+v", got)
	}
}

func TestProfileSetVisibility(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	svc := NewProfileService(users)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	got, err := svc.SetVisibility(ctx, u.ID(), entity.VisibilityPrivate)
	if err != nil {
		t.Fatalf("SetVisibility err: %v", err)
	}
	if got.Visibility() != entity.VisibilityPrivate {
		t.Fatalf("visibility = %v, want private", got.Visibility())
	}
}

func TestProfileSetAllowStrangerAdd(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	svc := NewProfileService(users)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	got, err := svc.SetAllowStrangerAdd(ctx, u.ID(), false)
	if err != nil {
		t.Fatalf("SetAllowStrangerAdd err: %v", err)
	}
	if got.AllowStrangerAdd() {
		t.Fatal("allow_stranger_add should be false")
	}
}

func TestProfileSetLocale(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	svc := NewProfileService(users)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u

	got, err := svc.SetLocale(ctx, u.ID(), "en-US")
	if err != nil {
		t.Fatalf("SetLocale err: %v", err)
	}
	if got.Locale() != "en-US" {
		t.Fatalf("locale = %q, want en-US", got.Locale())
	}
}

func TestProfileFindByIDNotFound(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo() // 空，FindByID 返回 ErrNotFound
	svc := NewProfileService(users)

	uid := idVal(t, "u-nonexistent")
	if _, err := svc.UpdateProfile(ctx, uid, "x", "", ""); err == nil {
		t.Fatal("UpdateProfile should error for nonexistent user")
	}
	if _, err := svc.SetVisibility(ctx, uid, entity.VisibilityPrivate); err == nil {
		t.Fatal("SetVisibility should error for nonexistent user")
	}
	if _, err := svc.SetAllowStrangerAdd(ctx, uid, false); err == nil {
		t.Fatal("SetAllowStrangerAdd should error for nonexistent user")
	}
	if _, err := svc.SetLocale(ctx, uid, "en-US"); err == nil {
		t.Fatal("SetLocale should error for nonexistent user")
	}
}
