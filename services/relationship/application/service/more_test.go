package service

import (
	"context"
	"testing"

	"isletalk/relationship/domain/entity"
)

func TestSetRemark(t *testing.T) {
	ctx := context.Background()
	remarks := newMemRemarks()
	svc := NewRemarkService(remarks)

	remark, err := svc.SetRemark(ctx, mustUID(t, "a"), mustUID(t, "b"), "best friend")
	if err != nil {
		t.Fatalf("SetRemark err: %v", err)
	}
	if remark.Content().String() != "best friend" {
		t.Fatal("content wrong")
	}
}

func TestDeleteFriendBidirectional(t *testing.T) {
	ctx := context.Background()
	friendships := newMemFriendships()
	requests := newMemRequests()
	identity := newMemIdentity()
	identity.allow["b"] = true
	fsvc := NewFriendRequestService(requests, identity)
	rsvc := NewRespondService(requests, friendships)
	lsvc := NewFriendListService(friendships)

	if _, err := fsvc.Send(ctx, mustUID(t, "a"), mustUID(t, "b"), "hi"); err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Respond(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatal(err)
	}

	if err := lsvc.Delete(ctx, mustUID(t, "a"), mustUID(t, "b")); err != nil {
		t.Fatalf("Delete err: %v", err)
	}
	// 双向解除
	if f, _ := friendships.Find(ctx, mustUID(t, "a"), mustUID(t, "b")); f.IsActive() {
		t.Fatal("friendship A→B should be deleted")
	}
	if f, _ := friendships.Find(ctx, mustUID(t, "b"), mustUID(t, "a")); f.IsActive() {
		t.Fatal("friendship B→A should be deleted")
	}
}

func TestListFriendsPagination(t *testing.T) {
	ctx := context.Background()
	friendships := newMemFriendships()
	svc := NewFriendListService(friendships)

	// 预置 3 个好友
	for _, id := range []string{"b", "c", "d"} {
		friendships.Create(ctx, entity.NewFriendship(mustUID(t, "a"), mustUID(t, id)))
	}

	got, err := svc.List(ctx, mustUID(t, "a"), 0, 2)
	if err != nil {
		t.Fatalf("List err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("page size = %d, want 2", len(got))
	}
}

func TestSetBlockUnblock(t *testing.T) {
	ctx := context.Background()
	blocks := newMemBlocks()
	svc := NewBlockService(blocks)

	if err := svc.SetBlock(ctx, mustUID(t, "a"), mustUID(t, "b"), true); err != nil {
		t.Fatalf("SetBlock err: %v", err)
	}
	if b, _ := blocks.Find(ctx, mustUID(t, "a"), mustUID(t, "b")); !b.IsActive() {
		t.Fatal("block should be active")
	}
	if err := svc.SetBlock(ctx, mustUID(t, "a"), mustUID(t, "b"), false); err != nil {
		t.Fatalf("unblock err: %v", err)
	}
	if b, _ := blocks.Find(ctx, mustUID(t, "a"), mustUID(t, "b")); b.IsActive() {
		t.Fatal("block should be unblocked")
	}
}
