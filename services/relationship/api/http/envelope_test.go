package http

import (
	"testing"

	"isletalk/relationship/application/service"
)

func TestErrorCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{service.ErrSelfRequest, "ARG_INVALID"},
		{service.ErrStrangerAddDisabled, "RES_STRANGER_ADD_DISABLED"},
		{service.ErrRequestExists, "RES_REQUEST_EXISTS"},
		{service.ErrRequestNotFound, "RES_NOT_FOUND"},
		{service.ErrBlocked, "RES_BLOCKED"},
		{service.ErrAlreadyFriends, "RES_ALREADY_FRIENDS"},
	}
	for _, c := range cases {
		code, _ := errorCode(c.err)
		if code != c.want {
			t.Fatalf("errorCode(%v) = %q, want %q", c.err, code, c.want)
		}
	}
}

func TestErrorCodeDefault(t *testing.T) {
	code, _ := errorCode(errUnknown)
	if code != "SYS_UNAVAILABLE" {
		t.Fatalf("default code = %q, want SYS_UNAVAILABLE", code)
	}
}

func TestErrorEnvelope(t *testing.T) {
	env := errorEnvelope(service.ErrBlocked)
	if env.Meta == nil || env.Meta.Code != "RES_BLOCKED" {
		t.Fatalf("envelope meta = %v, want RES_BLOCKED", env.Meta)
	}
	if env.Success {
		t.Fatal("error envelope should have success=false")
	}
}
