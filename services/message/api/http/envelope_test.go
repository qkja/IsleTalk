package http

import (
	"errors"
	"testing"

	"isletalk/message/application/service"
)

func TestErrorCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{service.ErrNotFriends, "MSG_NOT_FRIENDS"},
		{service.ErrBlocked, "MSG_BLOCKED"},
		{service.ErrMessageNotFound, "MSG_NOT_FOUND"},
		{service.ErrConversationNotFound, "CONV_NOT_FOUND"},
		{service.ErrRecallWindowExpired, "MSG_RECALL_WINDOW_EXPIRED"},
		{service.ErrNotSender, "MSG_NOT_SENDER"},
		{service.ErrInvalidMessageType, "ARG_INVALID"},
		{service.ErrNotParticipant, "MSG_NOT_PARTICIPANT"},
		{errors.New("unknown"), "SYS_UNAVAILABLE"},
	}
	for _, tc := range cases {
		code, msg := errorCode(tc.err)
		if code != tc.code || msg == "" {
			t.Fatalf("errorCode(%v) = %q, %q", tc.err, code, msg)
		}
	}
}

func TestEnvelopeHelpers(t *testing.T) {
	if e := ok("data"); !e.Success || e.Data != "data" {
		t.Fatalf("ok envelope wrong: %+v", e)
	}
	e := fail("CODE", "msg")
	if e.Success || e.Meta == nil || e.Meta.Code != "CODE" || e.Message == nil || *e.Message != "msg" {
		t.Fatalf("fail envelope wrong: %+v", e)
	}
	ee := errorEnvelope(service.ErrNotFriends)
	if ee.Success || ee.Meta == nil || ee.Meta.Code != "MSG_NOT_FRIENDS" {
		t.Fatalf("errorEnvelope wrong: %+v", ee)
	}
}
