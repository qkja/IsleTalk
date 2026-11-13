package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"isletalk/message/domain/entity"
)

func TestListMessagesPagination(t *testing.T) {
	msgs := newMemMessageRepo()
	svc := NewListMessagesService(msgs)
	ctx := context.Background()
	convID := "direct:a:b"
	a := uid(t, "a")
	for i := 1; i <= 5; i++ {
		m := entity.NewMessage(mid(t, fmt.Sprintf("m%d", i)), convID, a, entity.MessageText, content(t, "x"), uint64(i))
		msgs.byID[m.MessageID().String()] = m
	}

	got, next, err := svc.List(ctx, convID, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 || next != "" {
		t.Fatalf("got %d messages, next %q", len(got), next)
	}
	if got[0].Seq() != 5 {
		t.Fatalf("first seq = %d, want 5 (descending)", got[0].Seq())
	}
}

func TestListMessagesPageSizeCap(t *testing.T) {
	msgs := newMemMessageRepo()
	svc := NewListMessagesService(msgs)
	ctx := context.Background()
	convID := "direct:a:b"
	a := uid(t, "a")
	for i := 1; i <= 3; i++ {
		m := entity.NewMessage(mid(t, fmt.Sprintf("m%d", i)), convID, a, entity.MessageText, content(t, "x"), uint64(i))
		msgs.byID[m.MessageID().String()] = m
	}

	if _, _, err := svc.List(ctx, convID, 10000, ""); err != nil {
		t.Fatal(err)
	}
}

func TestListMessagesCursor(t *testing.T) {
	msgs := newMemMessageRepo()
	svc := NewListMessagesService(msgs)
	ctx := context.Background()
	convID := "direct:a:b"
	a := uid(t, "a")
	for i := 1; i <= 3; i++ {
		m := entity.NewMessage(mid(t, fmt.Sprintf("m%d", i)), convID, a, entity.MessageText, content(t, "x"), uint64(i))
		msgs.byID[m.MessageID().String()] = m
	}

	got, next, err := svc.List(ctx, convID, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || next == "" {
		t.Fatalf("got %d, next %q", len(got), next)
	}
	got2, next2, err := svc.List(ctx, convID, 2, next)
	if err != nil {
		t.Fatal(err)
	}
	if len(got2) != 1 || next2 != "" {
		t.Fatalf("got2 %d, next2 %q", len(got2), next2)
	}
}

func TestListMessagesInvalidToken(t *testing.T) {
	svc := NewListMessagesService(newMemMessageRepo())
	_, _, err := svc.List(context.Background(), "direct:a:b", 10, "not-a-number")
	if !errors.Is(err, ErrInvalidPageToken) {
		t.Fatalf("err = %v, want ErrInvalidPageToken", err)
	}
}
