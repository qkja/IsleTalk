package service

import (
	"context"
	"testing"
)

func TestSendFriendRequestSuccess(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	identity := newMemIdentity()
	identity.allow["b"] = true
	svc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)

	req, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hello")
	if err != nil {
		t.Fatalf("Send err: %v", err)
	}
	if !req.IsPending() {
		t.Fatal("request should be pending")
	}
}

func TestSendFriendRequestStrangerAddDisabled(t *testing.T) {
	ctx := context.Background()
	identity := newMemIdentity()
	identity.allow["b"] = false
	svc := NewFriendRequestService(newMemRequests(), newMemFriendships(), newMemBlocks(), identity, nil)

	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != ErrStrangerAddDisabled {
		t.Fatalf("err = %v, want ErrStrangerAddDisabled", err)
	}
}

func TestSendFriendRequestDuplicate(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	identity := newMemIdentity()
	identity.allow["b"] = true
	svc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)

	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatalf("first Send err: %v", err)
	}
	if _, err := svc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "again"); err != ErrRequestExists {
		t.Fatalf("err = %v, want ErrRequestExists", err)
	}
}

func TestRespondAcceptBidirectional(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	friendships := newMemFriendships()
	identity := newMemIdentity()
	identity.allow["b"] = true
	fsvc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)
	rsvc := NewRespondService(requests, friendships)

	if _, err := fsvc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatalf("Send err: %v", err)
	}
	if _, err := rsvc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatalf("Respond err: %v", err)
	}

	// 双向一致：A→B 与 B→A 都存在且 active（SC-002）
	if f, err := friendships.Find(ctx, mustUID(t, "a"), mustUID(t, "b")); err != nil || !f.IsActive() {
		t.Fatal("friendship A→B should exist and be active")
	}
	if f, err := friendships.Find(ctx, mustUID(t, "b"), mustUID(t, "a")); err != nil || !f.IsActive() {
		t.Fatal("friendship B→A should exist and be active")
	}
}

func TestRespondReject(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	friendships := newMemFriendships()
	identity := newMemIdentity()
	identity.allow["b"] = true
	fsvc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)
	rsvc := NewRespondService(requests, friendships)

	if _, err := fsvc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatalf("Send err: %v", err)
	}
	req, err := rsvc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), false)
	if err != nil {
		t.Fatalf("Respond err: %v", err)
	}
	if req.IsPending() {
		t.Fatal("request should be rejected")
	}
	if _, err := friendships.Find(ctx, mustUID(t, "a"), mustUID(t, "b")); err == nil {
		t.Fatal("reject should not create friendship")
	}
}

func TestCanMessageNotFriends(t *testing.T) {
	ctx := context.Background()
	csvc := NewCanMessageService(newMemFriendships(), newMemBlocks())
	allowed, blocked, _ := csvc.CanMessage(ctx, mustUID(t, "a"), mustUID(t, "b"))
	if allowed || blocked {
		t.Fatal("non-friends: allowed should be false, blocked should be false (FR-013 区分非好友与被拉黑)")
	}
}

func TestCanMessageAllowed(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	friendships := newMemFriendships()
	blocks := newMemBlocks()
	identity := newMemIdentity()
	identity.allow["b"] = true
	fsvc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)
	rsvc := NewRespondService(requests, friendships)
	csvc := NewCanMessageService(friendships, blocks)

	if _, err := fsvc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatal(err)
	}
	allowed, _, err := csvc.CanMessage(ctx, mustUID(t, "a"), mustUID(t, "b"))
	if err != nil || !allowed {
		t.Fatalf("CanMessage = %v, %v, want true", allowed, err)
	}
}

func TestCanMessageBlocked(t *testing.T) {
	ctx := context.Background()
	requests := newMemRequests()
	friendships := newMemFriendships()
	blocks := newMemBlocks()
	identity := newMemIdentity()
	identity.allow["b"] = true
	fsvc := NewFriendRequestService(requests, newMemFriendships(), newMemBlocks(), identity, nil)
	rsvc := NewRespondService(requests, friendships)
	bsvc := NewBlockService(blocks, requests)
	csvc := NewCanMessageService(friendships, blocks)

	if _, err := fsvc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatal(err)
	}
	// A 拉黑 B
	if err := bsvc.SetBlock(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatal(err)
	}
	allowed, blocked, _ := csvc.CanMessage(ctx, mustUID(t, "a"), mustUID(t, "b"))
	if allowed {
		t.Fatal("CanMessage should be false after block (FR-009)")
	}
	if !blocked {
		t.Fatal("blocked should be true after block (FR-013 原因类别)")
	}
}
