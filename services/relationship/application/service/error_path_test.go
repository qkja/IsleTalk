package service

import (
	"context"
	"testing"

	"isletalk/relationship/domain/entity"
)

func TestSendSelfRequest(t *testing.T) {
	ctx := context.Background()
	svc := NewFriendRequestService(newMemRequests(), newMemFriendships(), newMemBlocks(), newMemIdentity(), nil)
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "a"), "hi"); err != ErrSelfRequest {
		t.Fatalf("err = %v, want ErrSelfRequest", err)
	}
}

func TestSendIdentityFails(t *testing.T) {
	ctx := context.Background()
	identity := newMemIdentity() // 无 "b"
	svc := NewFriendRequestService(newMemRequests(), newMemFriendships(), newMemBlocks(), identity, nil)
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

func TestSendBlockedByTarget(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	blocks := newMemBlocks()
	identity := newMemIdentity()
	identity.allow["a"] = true
	// 目标 B 拉黑了发起方 A
	if err := blocks.Set(ctx, mustUID(t, "b"), mustUID(t, "a"), true); err != nil {
		t.Fatal(err)
	}
	svc := NewFriendRequestService(requests, newMemFriendships(), blocks, identity, nil)
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != ErrBlocked {
		t.Fatalf("err = %v, want ErrBlocked (FR-009)", err)
	}
}

func TestSendAlreadyFriends(t *testing.T) {
	ctx := context.Background()
	friendships := newMemFriendships()
	friendships.Create(ctx, entity.NewFriendship(mustUID(t, "a"), mustUID(t, "b")))
	svc := NewFriendRequestService(newMemRequests(), friendships, newMemBlocks(), newMemIdentity(), nil)
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != ErrAlreadyFriends {
		t.Fatalf("err = %v, want ErrAlreadyFriends", err)
	}
}

func TestBlockRejectsPendingRequest(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	blocks := newMemBlocks()
	// B 发申请给 A
	requests.Create(ctx, entity.NewFriendRequest(mustUID(t, "b"), mustUID(t, "a"), "hi"))
	svc := NewBlockService(blocks, requests)

	// A 拉黑 B → B 发来的待处理申请失效
	if err := svc.SetBlock(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatal(err)
	}
	req, _ := requests.Find(ctx, mustUID(t, "b"), mustUID(t, "a"))
	if req.IsPending() {
		t.Fatal("pending request should be rejected after block")
	}
}

func TestListRequests(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	requests.Create(ctx, entity.NewFriendRequest(mustUID(t, "b"), mustUID(t, "a"), "hi"))
	svc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), newMemIdentity(), nil)

	pending, err := svc.ListPending(ctx, mustUID(t, "a"))
	if err != nil || len(pending) != 1 {
		t.Fatalf("ListPending = %d, %v, want 1", len(pending), err)
	}
	history, err := svc.ListHistory(ctx, mustUID(t, "a"))
	if err != nil || len(history) != 0 {
		t.Fatalf("ListHistory = %d, %v, want 0", len(history), err)
	}
}
