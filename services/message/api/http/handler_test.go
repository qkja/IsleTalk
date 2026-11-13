package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"isletalk/message/application/service"
	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
)

type memConvs struct {
	m map[string]entity.Conversation
}

func (r *memConvs) Save(_ context.Context, c entity.Conversation) (entity.Conversation, error) {
	r.m[c.ConversationID()] = c
	return c, nil
}
func (r *memConvs) FindByID(_ context.Context, id string) (entity.Conversation, error) {
	if c, ok := r.m[id]; ok {
		return c, nil
	}
	return entity.Conversation{}, port.ErrNotFound
}
func (r *memConvs) ListByUser(_ context.Context, uid valueobject.UserId, limit int) ([]entity.Conversation, error) {
	out := []entity.Conversation{}
	for _, c := range r.m {
		if c.HasParticipant(uid) {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *memConvs) NextSeq(_ context.Context, _ string) (uint64, error) { return 0, nil }
func (r *memConvs) TouchOnSend(_ context.Context, _, _ string) error    { return nil }
func (r *memConvs) ResetUnread(_ context.Context, _ string) error       { return nil }

func (r *memConvs) TouchOnRecall(_ context.Context, _, _ string) error { return nil }

func newHandler() (*Handler, *memConvs) {
	convs := &memConvs{m: map[string]entity.Conversation{}}
	a, _ := valueobject.NewUserId("a")
	b, _ := valueobject.NewUserId("b")
	convs.m["direct:a:b"] = entity.NewConversation("direct:a:b", []valueobject.UserId{a, b})
	return NewHandler(service.NewConversationService(convs, nil, nil)), convs
}

func TestHandlerHealthz(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.Healthz(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandlerListConversations(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/conversations?userId=a", nil)
	w := httptest.NewRecorder()
	h.ListConversations(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var env Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if !env.Success {
		t.Fatalf("envelope should succeed: %+v", env)
	}
}

func TestHandlerListConversationsBadUser(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/conversations?userId=", nil)
	w := httptest.NewRecorder()
	h.ListConversations(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestHandlerPin(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/pin",
		strings.NewReader(`{"conversationId":"direct:a:b","userId":"a","pinned":true}`))
	w := httptest.NewRecorder()
	h.Pin(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
}

func TestHandlerMute(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/mute",
		strings.NewReader(`{"conversationId":"direct:a:b","userId":"a","muted":true}`))
	w := httptest.NewRecorder()
	h.Mute(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandlerMarkUnread(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/unread",
		strings.NewReader(`{"conversationId":"direct:a:b","userId":"a"}`))
	w := httptest.NewRecorder()
	h.MarkUnread(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandlerDelete(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/conversations",
		strings.NewReader(`{"conversationId":"direct:a:b","userId":"a"}`))
	w := httptest.NewRecorder()
	h.Delete(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestHandlerPinBadJSON(t *testing.T) {
	h, _ := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/pin", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	h.Pin(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
