package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

func newRecallFixture() (*RecallService, *memMessageRepo, *memConversationRepo) {
	msgs := newMemMessageRepo()
	convs := newMemConversationRepo()
	return NewRecallService(msgs, convs, nil), msgs, convs
}

func TestRecallWithinWindow(t *testing.T) {
	svc, msgs, _ := newRecallFixture()
	ctx := context.Background()
	a := uid(t, "a")
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), "direct:a:b", a, entity.MessageText, content(t, "hi"), 1)

	ok, err := svc.Recall(ctx, mid(t, "m1"), a)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if !ok {
		t.Fatal("recall should succeed within window")
	}
	got, _ := msgs.FindByID(ctx, mid(t, "m1"))
	if !got.Recalled() {
		t.Fatal("message should be recalled")
	}
}

func TestRecallUpdatesLatestPreview(t *testing.T) {
	svc, msgs, convs := newRecallFixture()
	ctx := context.Background()
	a := uid(t, "a")
	convID := "direct:a:b"
	convs.byID[convID] = entity.NewConversation(convID, []valueobject.UserId{a, uid(t, "b")})
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), convID, a, entity.MessageText, content(t, "secret"), 1)

	if _, err := svc.Recall(ctx, mid(t, "m1"), a); err != nil {
		t.Fatal(err)
	}
	c, _ := convs.FindByID(ctx, convID)
	if c.LastMessagePreview() != RecallPreview {
		t.Fatalf("preview = %q, want %q", c.LastMessagePreview(), RecallPreview)
	}
}

func TestRecallNotLatestKeepsPreview(t *testing.T) {
	svc, msgs, convs := newRecallFixture()
	ctx := context.Background()
	a := uid(t, "a")
	convID := "direct:a:b"
	c := entity.NewConversation(convID, []valueobject.UserId{a, uid(t, "b")})
	c.SetPreview("latest")
	convs.byID[convID] = c
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), convID, a, entity.MessageText, content(t, "old"), 1)
	msgs.byID["m2"] = entity.NewMessage(mid(t, "m2"), convID, a, entity.MessageText, content(t, "latest"), 2)

	if _, err := svc.Recall(ctx, mid(t, "m1"), a); err != nil {
		t.Fatal(err)
	}
	c, _ = convs.FindByID(ctx, convID)
	if c.LastMessagePreview() != "latest" {
		t.Fatalf("preview = %q, want unchanged latest", c.LastMessagePreview())
	}
}

func TestRecallWindowExpired(t *testing.T) {
	svc, msgs, _ := newRecallFixture()
	a := uid(t, "a")
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), "direct:a:b", a, entity.MessageText, content(t, "hi"), 1)
	svc.now = func() time.Time { return time.Now().Add(3 * time.Minute) }
	_, err := svc.Recall(context.Background(), mid(t, "m1"), a)
	if !errors.Is(err, ErrRecallWindowExpired) {
		t.Fatalf("err = %v, want ErrRecallWindowExpired", err)
	}
}

func TestRecallNotSender(t *testing.T) {
	svc, msgs, _ := newRecallFixture()
	a, b := uid(t, "a"), uid(t, "b")
	msgs.byID["m1"] = entity.NewMessage(mid(t, "m1"), "direct:a:b", a, entity.MessageText, content(t, "hi"), 1)
	_, err := svc.Recall(context.Background(), mid(t, "m1"), b)
	if !errors.Is(err, ErrNotSender) {
		t.Fatalf("err = %v, want ErrNotSender", err)
	}
}

func TestRecallNotFound(t *testing.T) {
	svc, _, _ := newRecallFixture()
	_, err := svc.Recall(context.Background(), mid(t, "missing"), uid(t, "a"))
	if !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("err = %v, want ErrMessageNotFound", err)
	}
}
