package entity

import (
	"testing"
	"time"

	"isletalk/identity/domain/valueobject"
)

// 覆盖所有 getter/setter（提升 domain/entity 覆盖率）。

func TestUserGetters(t *testing.T) {
	u, _ := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "zh-CN")

	if u.ID().String() != "id-1" {
		t.Fatal("ID")
	}
	if u.PublicName().String() != "alice" {
		t.Fatal("PublicName")
	}
	if u.PhoneHash().String() != "phone-1" {
		t.Fatal("PhoneHash")
	}
	if u.Nickname() != "" {
		t.Fatal("Nickname should default empty")
	}
	if u.AvatarURL() != "" {
		t.Fatal("AvatarURL should default empty")
	}
	if u.Signature() != "" {
		t.Fatal("Signature should default empty")
	}
	if u.Visibility() != VisibilityFriendsOnly {
		t.Fatalf("default visibility = %v", u.Visibility())
	}
	if !u.AllowStrangerAdd() {
		t.Fatal("default allow_stranger_add should be true")
	}
	if u.Locale() != "zh-CN" {
		t.Fatalf("locale = %q", u.Locale())
	}
	if u.Status() != UserActive {
		t.Fatal("status should be active")
	}
	if u.CreatedAt().IsZero() {
		t.Fatal("created_at should be set")
	}
}

func TestUserSetters(t *testing.T) {
	u, _ := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "zh-CN")

	u.UpdateProfile("Alice", "http://a/1.png", "hi")
	u.SetVisibility(VisibilityPublic)
	u.SetAllowStrangerAdd(false)
	u.SetLocale("en-US")
	pw, _ := valueobject.NewPasswordHash("hash:x")
	u.SetPassword(pw)

	if u.Nickname() != "Alice" || u.AvatarURL() != "http://a/1.png" || u.Signature() != "hi" {
		t.Fatal("UpdateProfile")
	}
	if u.Visibility() != VisibilityPublic {
		t.Fatal("SetVisibility")
	}
	if u.AllowStrangerAdd() {
		t.Fatal("SetAllowStrangerAdd")
	}
	if u.Locale() != "en-US" {
		t.Fatal("SetLocale")
	}
	if u.PasswordHash().String() != "hash:x" {
		t.Fatal("SetPassword")
	}
}

func TestUserInvalidVisibilityFallsBack(t *testing.T) {
	u, _ := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "zh-CN")
	u.SetVisibility(Visibility("bogus"))
	if u.Visibility() != VisibilityFriendsOnly {
		t.Fatalf("invalid visibility should fall back to friends_only, got %v", u.Visibility())
	}
}

func TestSessionGetters(t *testing.T) {
	uid := id(t, "u-1")
	s := NewSession("s-1", uid, "dev-1")

	if s.ID() != "s-1" || s.UserID() != uid || s.DeviceID() != "dev-1" {
		t.Fatal("session getters")
	}
	if s.CreatedAt().IsZero() || s.LastActiveAt().IsZero() {
		t.Fatal("timestamps should be set")
	}
	if s.Revoked() {
		t.Fatal("should not be revoked")
	}
	s.Touch()
	if s.LastActiveAt().Before(s.CreatedAt()) {
		t.Fatal("Touch should update last_active_at")
	}
}

func TestVerificationGetters(t *testing.T) {
	uid := id(t, "u-1")
	now := time.Now()
	v := NewVerification(uid, "idcard", "digest", now, now.Add(30*24*time.Hour))

	if v.UserID() != uid || v.IDType() != "idcard" || v.IDDigest() != "digest" {
		t.Fatal("verification getters")
	}
	if v.Status() != VerificationSubmitted {
		t.Fatal("status should be submitted")
	}
	if !v.ConsentAt().Equal(now) || !v.RetentionUntil().Equal(now.Add(30*24*time.Hour)) {
		t.Fatal("verification timestamps")
	}
}

func TestTrustLevelGettersAndDowngrade(t *testing.T) {
	uid := id(t, "u-1")
	tl := NewTrustLevel(uid, 2, map[string]int{"friends": 500})

	if tl.UserID() != uid || tl.Level() != 2 || tl.Quota()["friends"] != 500 {
		t.Fatal("trust level getters")
	}
	tl.Downgrade()
	if tl.Level() != 0 || len(tl.Quota()) != 0 {
		t.Fatal("downgrade should reset level and quota")
	}
}

func TestNameClaimGetters(t *testing.T) {
	uid := id(t, "u-1")
	n := name(t, "alice")
	c := NewNameClaim(n, uid)

	if c.Name() != n || c.OwnerUserID() != uid {
		t.Fatal("name claim getters")
	}
}
