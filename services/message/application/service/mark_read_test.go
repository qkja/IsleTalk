package service

import (
	"context"
	"fmt"
	"testing"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

func TestMarkRead(t *testing.T) {
	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	svc := NewMarkReadService(msgs, convs, nil, nil)
	ctx := context.Background()

	convID := "direct:a:b"
	a, b := uid(t, "a"), uid(t, "b")
	c := entity.NewConversation(convID, []valueobject.UserId{a, b})
	c.IncrementUnread()
	c.IncrementUnread()
	convs.byID[convID] = c

	for i := 1; i <= 2; i++ {
		m := entity.NewMessage(mid(t, fmt.Sprintf("m%d", i)), convID, a, entity.MessageText, content(t, "hi"), uint64(i))
		m.MarkDelivered()
		msgs.byID[m.MessageID().String()] = m
	}

	if err := svc.MarkRead(ctx, convID, b, 2); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	for _, m := range msgs.byID {
		if m.Status() != entity.MessageRead {
			t.Fatalf("message status = %v, want read", m.Status())
		}
	}
	got, _ := convs.FindByID(ctx, convID)
	if got.UnreadCount() != 0 {
		t.Fatalf("unread = %d, want 0", got.UnreadCount())
	}
}

func TestMarkReadDoesNotAffectOwnMessages(t *testing.T) {
	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	svc := NewMarkReadService(msgs, convs, nil, nil)
	ctx := context.Background()

	convID := "direct:a:b"
	a := uid(t, "a")
	m := entity.NewMessage(mid(t, "m1"), convID, a, entity.MessageText, content(t, "hi"), 1)
	m.MarkDelivered()
	msgs.byID["m1"] = m

	if err := svc.MarkRead(ctx, convID, a, 1); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	got, _ := msgs.FindByID(ctx, mid(t, "m1"))
	if got.Status() != entity.MessageDelivered {
		t.Fatalf("sender's own message should stay delivered, got %v", got.Status())
	}
}
