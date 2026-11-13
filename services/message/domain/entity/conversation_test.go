package entity

import (
	"testing"
	"time"

	"isletalk/message/domain/valueobject"
)

func TestNewConversationDefaults(t *testing.T) {
	a, b := mustUID(t, "a"), mustUID(t, "b")
	c := NewConversation("direct:a:b", []valueobject.UserId{a, b})

	if c.ConversationID() != "direct:a:b" || c.Type() != ConversationDirect {
		t.Fatalf("id/type mismatch: %+v", c)
	}
	if len(c.Participants()) != 2 || !c.HasParticipant(a) || !c.HasParticipant(b) {
		t.Fatal("participants should contain a and b")
	}
	if c.UnreadCount() != 0 || c.Pinned() || c.Muted() || c.Deleted() {
		t.Fatal("defaults wrong: unread should be 0, flags false")
	}
	if c.LastActiveAt().IsZero() {
		t.Fatal("last active at should be set")
	}
}

func TestDirectConversationIDDeterministic(t *testing.T) {
	a, b := mustUID(t, "a"), mustUID(t, "b")
	if DirectConversationID(a, b) != DirectConversationID(b, a) {
		t.Fatal("DirectConversationID should be order-independent")
	}
	if got := DirectConversationID(a, b); got != "direct:a:b" {
		t.Fatalf("DirectConversationID = %q, want direct:a:b", got)
	}
}

func TestParseDirectConversationID(t *testing.T) {
	a, b := mustUID(t, "a"), mustUID(t, "b")
	pa, pb, ok := ParseDirectConversationID("direct:a:b")
	if !ok || !pa.Equal(a) || !pb.Equal(b) {
		t.Fatalf("parse roundtrip failed: %v %v %v", pa, pb, ok)
	}

	for _, id := range []string{"", "group:a:b", "direct:", "direct:a", "direct:a:", "direct::b"} {
		if _, _, ok := ParseDirectConversationID(id); ok {
			t.Fatalf("ParseDirectConversationID(%q) should fail", id)
		}
	}
}

func TestConversationUnreadAndFlags(t *testing.T) {
	c := NewConversation("direct:a:b", []valueobject.UserId{mustUID(t, "a"), mustUID(t, "b")})

	c.IncrementUnread()
	c.IncrementUnread()
	if c.UnreadCount() != 2 {
		t.Fatalf("unread = %d, want 2", c.UnreadCount())
	}
	c.ResetUnread()
	if c.UnreadCount() != 0 {
		t.Fatalf("unread = %d, want 0 after reset", c.UnreadCount())
	}

	c.Pin()
	if !c.Pinned() {
		t.Fatal("Pin should set pinned")
	}
	c.Unpin()
	if c.Pinned() {
		t.Fatal("Unpin should clear pinned")
	}

	c.Mute()
	if !c.Muted() {
		t.Fatal("Mute should set muted")
	}
	c.Unmute()
	if c.Muted() {
		t.Fatal("Unmute should clear muted")
	}

	c.MarkDeleted()
	if !c.Deleted() {
		t.Fatal("MarkDeleted should set deleted")
	}
	c.Restore()
	if c.Deleted() {
		t.Fatal("Restore should clear deleted")
	}
}

func TestConversationSetPreview(t *testing.T) {
	c := NewConversation("direct:a:b", []valueobject.UserId{mustUID(t, "a"), mustUID(t, "b")})
	before := c.LastActiveAt()
	time.Sleep(time.Millisecond)
	c.SetPreview("hello")
	if c.LastMessagePreview() != "hello" {
		t.Fatalf("preview = %q, want hello", c.LastMessagePreview())
	}
	if !c.LastActiveAt().After(before) {
		t.Fatal("SetPreview should update last active at")
	}
}

func TestReconstructConversation(t *testing.T) {
	at := time.UnixMilli(1700000000000)
	c := ReconstructConversation(
		"direct:a:b", ConversationDirect, []valueobject.UserId{mustUID(t, "a")},
		"preview", 3, true, true, true, at,
	)
	if c.ConversationID() != "direct:a:b" || c.Type() != ConversationDirect ||
		c.LastMessagePreview() != "preview" || c.UnreadCount() != 3 ||
		!c.Pinned() || !c.Muted() || !c.Deleted() || !c.LastActiveAt().Equal(at) {
		t.Fatalf("ReconstructConversation roundtrip failed: %+v", c)
	}
}
