package service

import (
	"context"
	"errors"
	"testing"

	"isletalk/message/domain/entity"
)

func newSendService(allowed, blocked bool) (*SendMessageService, *memMessageRepo, *memConversationRepo, *mockRelationship) {
	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	rel := &mockRelationship{allowed: allowed, blocked: blocked}
	svc := NewSendMessageService(msgs, convs, rel, nil, nil)
	return svc, msgs, convs, rel
}

func TestSendMessageSuccess(t *testing.T) {
	svc, msgs, convs, rel := newSendService(true, false)
	ctx := context.Background()
	a := uid(t, "a")
	convID := entity.DirectConversationID(a, uid(t, "b"))

	m, err := svc.Send(ctx, mid(t, "m1"), convID, a, entity.MessageText, content(t, "hi"))
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if m.Status() != entity.MessageDelivered {
		t.Fatalf("status = %v, want delivered", m.Status())
	}
	if m.Seq() != 1 {
		t.Fatalf("seq = %d, want 1", m.Seq())
	}
	if rel.calls != 1 {
		t.Fatalf("CanMessage calls = %d, want 1", rel.calls)
	}
	if _, err := convs.FindByID(ctx, convID); err != nil {
		t.Fatalf("conversation should exist: %v", err)
	}
	if len(msgs.byID) != 1 {
		t.Fatalf("messages = %d, want 1", len(msgs.byID))
	}
}

func TestSendMessageIdempotent(t *testing.T) {
	svc, msgs, _, _ := newSendService(true, false)
	ctx := context.Background()
	a := uid(t, "a")
	convID := entity.DirectConversationID(a, uid(t, "b"))
	msgID := mid(t, "m1")

	if _, err := svc.Send(ctx, msgID, convID, a, entity.MessageText, content(t, "hi")); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Send(ctx, msgID, convID, a, entity.MessageText, content(t, "hi")); err != nil {
		t.Fatal(err)
	}
	if len(msgs.byID) != 1 {
		t.Fatalf("idempotent send should not duplicate, got %d messages", len(msgs.byID))
	}
}

func TestSendMessageNotFriends(t *testing.T) {
	svc, _, _, _ := newSendService(false, false)
	a := uid(t, "a")
	convID := entity.DirectConversationID(a, uid(t, "b"))
	_, err := svc.Send(context.Background(), mid(t, "m1"), convID, a, entity.MessageText, content(t, "hi"))
	if !errors.Is(err, ErrNotFriends) {
		t.Fatalf("err = %v, want ErrNotFriends", err)
	}
}

func TestSendMessageBlocked(t *testing.T) {
	svc, _, _, _ := newSendService(false, true)
	a := uid(t, "a")
	convID := entity.DirectConversationID(a, uid(t, "b"))
	_, err := svc.Send(context.Background(), mid(t, "m1"), convID, a, entity.MessageText, content(t, "hi"))
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("err = %v, want ErrBlocked", err)
	}
}

func TestSendMessageInvalidType(t *testing.T) {
	svc, _, _, _ := newSendService(true, false)
	a := uid(t, "a")
	convID := entity.DirectConversationID(a, uid(t, "b"))
	_, err := svc.Send(context.Background(), mid(t, "m1"), convID, a, entity.MessageType("bad"), content(t, "x"))
	if !errors.Is(err, ErrInvalidMessageType) {
		t.Fatalf("err = %v, want ErrInvalidMessageType", err)
	}
}

func TestSendMessageNotParticipant(t *testing.T) {
	svc, _, _, _ := newSendService(true, false)
	convID := entity.DirectConversationID(uid(t, "a"), uid(t, "b"))
	_, err := svc.Send(context.Background(), mid(t, "m1"), convID, uid(t, "c"), entity.MessageText, content(t, "x"))
	if !errors.Is(err, ErrNotParticipant) {
		t.Fatalf("err = %v, want ErrNotParticipant", err)
	}
}

func TestSendMessageInvalidConversationID(t *testing.T) {
	svc, _, _, _ := newSendService(true, false)
	_, err := svc.Send(context.Background(), mid(t, "m1"), "not-a-conv", uid(t, "a"), entity.MessageText, content(t, "x"))
	if !errors.Is(err, ErrInvalidConversationID) {
		t.Fatalf("err = %v, want ErrInvalidConversationID", err)
	}
}
