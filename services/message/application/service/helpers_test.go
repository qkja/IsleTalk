package service

import (
	"testing"

	"isletalk/message/domain/valueobject"
)

func uid(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	u, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func mid(t *testing.T, s string) valueobject.MessageId {
	t.Helper()
	m, err := valueobject.NewMessageId(s)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func content(t *testing.T, s string) valueobject.MessageContent {
	t.Helper()
	c, err := valueobject.NewMessageContent(s)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
