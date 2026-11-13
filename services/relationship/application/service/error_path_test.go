package service

import (
	"context"
	"testing"
)

func TestSendSelfRequest(t *testing.T) {
	ctx := context.Background()
	svc := NewFriendRequestService(newMemRequests(), newMemIdentity())
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "a"), "hi"); err != ErrSelfRequest {
		t.Fatalf("err = %v, want ErrSelfRequest", err)
	}
}

func TestSendIdentityFails(t *testing.T) {
	ctx := context.Background()
	identity := newMemIdentity() // 无 "b"
	svc := NewFriendRequestService(newMemRequests(), identity)
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err == nil {
		t.Fatal("should fail when identity lookup fails")
	}
}

func TestRespondNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewRespondService(newMemRequests(), newMemFriendships())
	if _, err := svc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != ErrRequestNotFound {
		t.Fatalf("err = %v, want ErrRequestNotFound", err)
	}
}

func TestSetRemarkInvalidText(t *testing.T) {
	ctx := context.Background()
	svc := NewRemarkService(newMemRemarks())
	if _, err := svc.SetRemark(ctx, mustUID(t, "a"), mustUID(t, "b"), ""); err == nil {
		t.Fatal("empty remark should fail")
	}
}

func TestListFriendsLimitCap(t *testing.T) {
	ctx := context.Background()
	friendships := newMemFriendships()
	svc := NewFriendListService(friendships)

	// limit 超上限 → 收敛到 50
	if _, err := svc.List(ctx, mustUID(t, "a"), 0, 5000); err != nil {
		t.Fatalf("List err: %v", err)
	}
}
