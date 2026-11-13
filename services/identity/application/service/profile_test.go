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
