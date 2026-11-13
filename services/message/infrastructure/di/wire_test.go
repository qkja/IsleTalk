package di

import (
	"testing"

	"isletalk/message/application/service"
)

func TestNewContainer(t *testing.T) {
	c := NewContainer(nil, nil, nil)
	if c.SendMessage == nil || c.ListMessages == nil || c.Recall == nil || c.MarkRead == nil || c.Conversation == nil {
		t.Fatal("container should wire all services")
	}
	var _ *service.SendMessageService = c.SendMessage
	var _ *service.ListMessagesService = c.ListMessages
	var _ *service.RecallService = c.Recall
	var _ *service.MarkReadService = c.MarkRead
	var _ *service.ConversationService = c.Conversation
}
