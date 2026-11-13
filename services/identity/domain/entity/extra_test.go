package entity

import "testing"

func TestNewTrustLevelNilQuota(t *testing.T) {
	tl := NewTrustLevel(id(t, "id-1"), 1, nil)
	if tl.Level() != 1 {
		t.Fatalf("level = %d, want 1", tl.Level())
	}
	if tl.Quota() == nil {
		t.Fatal("nil quota should be defaulted to empty map")
	}
	if tl.UserID().String() != "id-1" {
		t.Fatalf("user id = %v", tl.UserID())
	}
}

func TestTrustLevelDowngrade(t *testing.T) {
	tl := NewTrustLevel(id(t, "id-1"), 3, map[string]int{"x": 10})
	tl.Downgrade()
	if tl.Level() != 0 {
		t.Fatalf("level = %d, want 0", tl.Level())
	}
	if len(tl.Quota()) != 0 {
		t.Fatal("downgrade should clear quota")
	}
}

func TestNewUserDefaultLocale(t *testing.T) {
	u, err := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "")
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}
	if u.Locale() != "zh-CN" {
		t.Fatalf("locale = %q, want zh-CN", u.Locale())
	}
	if u.Visibility() != VisibilityFriendsOnly || !u.AllowStrangerAdd() {
		t.Fatal("default visibility/stranger-add wrong")
	}
}
