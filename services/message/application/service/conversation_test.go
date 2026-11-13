package service

import (
	"context"
	"errors"
	"testing"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

func newConversationServiceFixture() (*ConversationService, *memConversationRepo) {
	convs := newMemConversationRepo()
	return NewConversationService(convs, nil, nil), convs
}

func seedConversation(t *testing.T, convs *memConversationRepo, convID string) {
	t.Helper()
	a, b := uid(t, "a"), uid(t, "b")
	convs.byID[convID] = entity.NewConversation(convID, []valueobject.UserId{a, b})
}

func TestConversationPin(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	if err := svc.Pin(ctx, "direct:a:b", uid(t, "a"), true); err != nil {
		t.Fatal(err)
	}
	c, _ := convs.FindByID(ctx, "direct:a:b")
	if !c.Pinned() {
		t.Fatal("conversation should be pinned")
	}
	if err := svc.Pin(ctx, "direct:a:b", uid(t, "a"), false); err != nil {
		t.Fatal(err)
	}
	c, _ = convs.FindByID(ctx, "direct:a:b")
	if c.Pinned() {
		t.Fatal("conversation should be unpinned")
	}
}

func TestConversationMute(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	if err := svc.Mute(ctx, "direct:a:b", uid(t, "a"), true); err != nil {
		t.Fatal(err)
	}
	c, _ := convs.FindByID(ctx, "direct:a:b")
	if !c.Muted() {
		t.Fatal("conversation should be muted")
	}
}

func TestConversationDelete(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	if err := svc.Delete(ctx, "direct:a:b", uid(t, "a")); err != nil {
		t.Fatal(err)
	}
	c, _ := convs.FindByID(ctx, "direct:a:b")
	if !c.Deleted() {
		t.Fatal("conversation should be deleted")
	}
}

func TestConversationMarkUnread(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	if err := svc.MarkUnread(ctx, "direct:a:b", uid(t, "a")); err != nil {
		t.Fatal(err)
	}
	c, _ := convs.FindByID(ctx, "direct:a:b")
	if c.UnreadCount() != 1 {
		t.Fatalf("unread = %d, want 1", c.UnreadCount())
	}
}

func TestConversationNotParticipant(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	err := svc.Pin(ctx, "direct:a:b", uid(t, "c"), true)
	if !errors.Is(err, ErrNotParticipant) {
		t.Fatalf("err = %v, want ErrNotParticipant", err)
	}
}

func TestConversationNotFound(t *testing.T) {
	svc, _ := newConversationServiceFixture()
	err := svc.Delete(context.Background(), "direct:missing", uid(t, "a"))
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("err = %v, want ErrConversationNotFound", err)
	}
}

func TestConversationList(t *testing.T) {
	svc, convs := newConversationServiceFixture()
	ctx := context.Background()
	seedConversation(t, convs, "direct:a:b")

	list, err := svc.List(ctx, uid(t, "a"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("list = %d, want 1", len(list))
	}
}
