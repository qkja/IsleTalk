package entity

import (
	"testing"

	"isletalk/relationship/domain/valueobject"
)

func uid(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestFriendshipLifecycle(t *testing.T) {
	f := NewFriendship(uid(t, "a"), uid(t, "b"))
	if !f.IsActive() {
		t.Fatal("new friendship should be active")
	}
	if f.UserID().String() != "a" || f.FriendID().String() != "b" {
		t.Fatal("friendship ids wrong")
	}
	f.Delete()
	if f.IsActive() {
		t.Fatal("friendship should be deleted")
	}
}

func TestFriendRequestLifecycle(t *testing.T) {
	r := NewFriendRequest(uid(t, "a"), uid(t, "b"), "hello")
	if !r.IsPending() {
		t.Fatal("new request should be pending")
	}
	if r.RequestMessage() != "hello" {
		t.Fatal("message wrong")
	}
	r.Accept()
	if r.IsPending() || r.Status() != RequestAccepted {
		t.Fatal("request should be accepted")
	}

	r2 := NewFriendRequest(uid(t, "a"), uid(t, "b"), "hi")
	r2.Reject()
	if r2.Status() != RequestRejected {
		t.Fatal("request should be rejected")
	}
}

func TestBlockLifecycle(t *testing.T) {
	b := NewBlock(uid(t, "a"), uid(t, "b"))
	if !b.IsActive() {
		t.Fatal("new block should be active")
	}
	if b.OwnerID().String() != "a" || b.TargetID().String() != "b" {
		t.Fatal("block ids wrong")
	}
	b.Unblock()
	if b.IsActive() {
		t.Fatal("block should be unblocked")
	}
}

func TestRemarkLifecycle(t *testing.T) {
	txt, _ := valueobject.NewRemarkText("best friend")
	r := NewRemark(uid(t, "a"), uid(t, "b"), txt)
	if r.Content().String() != "best friend" {
		t.Fatal("content wrong")
	}
	newTxt, _ := valueobject.NewRemarkText("colleague")
	r.Update(newTxt)
	if r.Content().String() != "colleague" {
		t.Fatal("update content wrong")
	}
}

func TestValueObjectValidation(t *testing.T) {
	if _, err := valueobject.NewUserId(""); err == nil {
		t.Fatal("empty user id should error")
	}
	if _, err := valueobject.NewRemarkText(""); err == nil {
		t.Fatal("empty remark should error")
	}
	if _, err := valueobject.NewRemarkText(string(make([]rune, 51))); err == nil {
		t.Fatal("overlong remark should error")
	}
}

func TestGetterCoverage(t *testing.T) {
	f := NewFriendship(uid(t, "a"), uid(t, "b"))
	if f.CreatedAt().IsZero() || f.Status() != FriendshipAccepted {
		t.Fatal("friendship getters wrong")
	}
	r := NewFriendRequest(uid(t, "a"), uid(t, "b"), "hi")
	if r.CreatedAt().IsZero() {
		t.Fatal("request CreatedAt zero")
	}
	b := NewBlock(uid(t, "a"), uid(t, "b"))
	if b.Status() != BlockActive {
		t.Fatal("block status wrong")
	}
}
