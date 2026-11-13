package entity

import (
	"testing"
	"time"

	"isletalk/message/domain/valueobject"
)

func mustUID(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	u, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func mustMID(t *testing.T, s string) valueobject.MessageId {
	t.Helper()
	m, err := valueobject.NewMessageId(s)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustContent(t *testing.T, s string) valueobject.MessageContent {
	t.Helper()
	c, err := valueobject.NewMessageContent(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMessageStateMachine(t *testing.T) {
	m := NewMessage(mustMID(t, "m1"), "conv1", mustUID(t, "a"), MessageText, mustContent(t, "hi"), 1)

	if m.Status() != MessageSent {
		t.Fatalf("initial status = %v, want sent", m.Status())
	}

	m.MarkDelivered()
	if m.Status() != MessageDelivered {
		t.Fatalf("after MarkDelivered status = %v, want delivered", m.Status())
	}

	m.MarkRead()
	if m.Status() != MessageRead {
		t.Fatalf("after MarkRead status = %v, want read", m.Status())
	}
}

func TestMessageStateMachineGuards(t *testing.T) {
	// sent 状态不能直接 read（必须先 delivered）
	m := NewMessage(mustMID(t, "m1"), "conv1", mustUID(t, "a"), MessageText, mustContent(t, "hi"), 1)
	m.MarkRead()
	if m.Status() != MessageSent {
		t.Fatalf("MarkRead on sent should be no-op, got %v", m.Status())
	}

	// MarkDelivered 幂等
	m.MarkDelivered()
	m.MarkDelivered()
	if m.Status() != MessageDelivered {
		t.Fatalf("double MarkDelivered should stay delivered, got %v", m.Status())
	}
}

func TestMessageFailed(t *testing.T) {
	m := NewMessage(mustMID(t, "m1"), "conv1", mustUID(t, "a"), MessageText, mustContent(t, "hi"), 1)
	m.MarkFailed()
	if m.Status() != MessageFailed {
		t.Fatalf("status = %v, want failed", m.Status())
	}
}

func TestMessageRecallAndSender(t *testing.T) {
	sender := mustUID(t, "a")
	m := NewMessage(mustMID(t, "m1"), "conv1", sender, MessageText, mustContent(t, "hi"), 1)

	if !m.IsSentBy(sender) {
		t.Fatal("IsSentBy should be true for sender")
	}
	if m.IsSentBy(mustUID(t, "b")) {
		t.Fatal("IsSentBy should be false for non-sender")
	}
	if m.Recalled() {
		t.Fatal("new message should not be recalled")
	}
	m.Recall()
	if !m.Recalled() {
		t.Fatal("after Recall, Recalled should be true")
	}
}

func TestMessageGetters(t *testing.T) {
	sentAt := time.UnixMilli(1700000000000)
	m := ReconstructMessage(
		mustMID(t, "m1"), "conv1", mustUID(t, "a"), MessageImage, mustContent(t, "img-key"),
		7, sentAt, MessageRead, true,
	)
	if m.MessageID().String() != "m1" || m.ConversationID() != "conv1" ||
		m.SenderID().String() != "a" || m.Type() != MessageImage ||
		m.Content().String() != "img-key" || m.Seq() != 7 ||
		!m.SentAt().Equal(sentAt) || m.Status() != MessageRead || !m.Recalled() {
		t.Fatalf("ReconstructMessage roundtrip failed: %+v", m)
	}
}
