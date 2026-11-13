package entity

import (
	"testing"
	"time"

	"isletalk/identity/domain/valueobject"
)

func id(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatalf("NewUserId: %v", err)
	}
	return v
}

func name(t *testing.T, s string) valueobject.PublicName {
	t.Helper()
	v, err := valueobject.NewPublicName(s)
	if err != nil {
		t.Fatalf("NewPublicName: %v", err)
	}
	return v
}

func phone(t *testing.T, s string) valueobject.PhoneHash {
	t.Helper()
	v, err := valueobject.NewPhoneHash(s)
	if err != nil {
		t.Fatalf("NewPhoneHash: %v", err)
	}
	return v
}

func TestUserSetPublicNameKeepsID(t *testing.T) {
	u, err := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "zh-CN")
	if err != nil {
		t.Fatalf("NewUser err: %v", err)
	}

	newName := name(t, "bob")
	u.SetPublicName(newName)

	if u.ID().String() != "id-1" {
		t.Fatalf("internal ID changed: %v", u.ID())
	}
	if u.PublicName() != newName {
		t.Fatalf("public name = %v, want %v", u.PublicName(), newName)
	}
}

func TestUserDeactivate(t *testing.T) {
	u, _ := NewUser(id(t, "id-1"), name(t, "alice"), phone(t, "phone-1"), "zh-CN")

	if !u.IsActive() {
		t.Fatal("new user should be active")
	}
	u.Deactivate()
	if u.IsActive() {
		t.Fatal("user should be deactivated")
	}
}

func TestSessionRevoke(t *testing.T) {
	s := NewSession("sess-1", id(t, "id-1"), "dev-1")

	if !s.IsActive() {
		t.Fatal("new session should be active")
	}
	s.Revoke()
	if s.IsActive() {
		t.Fatal("session should be revoked")
	}
}

func TestVerificationRevoke(t *testing.T) {
	v := NewVerification(id(t, "id-1"), "idcard", "encrypted-digest", time.Now(), time.Now().Add(30*24*time.Hour))

	if !v.IsSubmitted() {
		t.Fatal("new verification should be submitted")
	}
	v.Revoke()
	if v.IsSubmitted() {
		t.Fatal("verification should be revoked")
	}
}

func TestNameClaimReservation(t *testing.T) {
	claim := NewNameClaim(name(t, "alice"), id(t, "id-1"))

	now := time.Now()
	if claim.Available(now) {
		t.Fatal("occupied name should not be available")
	}

	claim.Release(now, 90*24*time.Hour)
	if claim.Available(now) {
		t.Fatal("released name within retention should not be available")
	}
	if !claim.Available(now.Add(91 * 24 * time.Hour)) {
		t.Fatal("name should be available after retention expires")
	}
}
