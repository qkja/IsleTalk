package valueobject

import (
	"strings"
	"testing"
)

func TestNewUserId(t *testing.T) {
	if _, err := NewUserId(""); err != ErrEmptyUserId {
		t.Fatalf("empty should return ErrEmptyUserId, got %v", err)
	}
	u, err := NewUserId("u1")
	if err != nil {
		t.Fatal(err)
	}
	if u.String() != "u1" {
		t.Fatalf("String() = %q, want u1", u.String())
	}
	other, _ := NewUserId("u1")
	if !u.Equal(other) {
		t.Fatal("Equal should be true for same value")
	}
	diff, _ := NewUserId("u2")
	if u.Equal(diff) {
		t.Fatal("Equal should be false for different value")
	}
}

func TestNewMessageId(t *testing.T) {
	if _, err := NewMessageId(""); err != ErrEmptyMessageId {
		t.Fatalf("empty should return ErrEmptyMessageId, got %v", err)
	}
	m, err := NewMessageId("m1")
	if err != nil {
		t.Fatal(err)
	}
	if m.String() != "m1" {
		t.Fatalf("String() = %q, want m1", m.String())
	}
	same, _ := NewMessageId("m1")
	if !m.Equal(same) {
		t.Fatal("Equal should be true for same value")
	}
}

func TestNewMessageContent(t *testing.T) {
	if _, err := NewMessageContent(""); err != ErrEmptyContent {
		t.Fatalf("empty should return ErrEmptyContent, got %v", err)
	}
	long := strings.Repeat("a", MaxContentBytes+1)
	if _, err := NewMessageContent(long); err != ErrContentTooLong {
		t.Fatalf("too long should return ErrContentTooLong, got %v", err)
	}
	c, err := NewMessageContent("hello")
	if err != nil {
		t.Fatal(err)
	}
	if c.String() != "hello" {
		t.Fatalf("String() = %q, want hello", c.String())
	}
}
