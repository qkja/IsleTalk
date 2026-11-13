package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var errUnknown = errors.New("unknown error")

func newTestHandler() *Handler {
	return NewHandler(nil, nil, nil, nil, nil)
}

func doRequest(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/healthz", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz code = %d, want 200", rec.Code)
	}
}

func TestSendRequestBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/friend-request", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestSendRequestInvalidUserID(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/friend-request", `{"initiatorId":"","targetId":"b"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestRespondBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/friend-respond", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestSetRemarkBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/remark", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestSetBlockBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/block", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestListFriendsInvalidUserID(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/api/v1/friends?userId=", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestDeleteFriendBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "DELETE", "/api/v1/friend", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}
