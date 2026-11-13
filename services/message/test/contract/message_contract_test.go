package contract

import (
	"context"
	"sort"
	"testing"

	grpcserver "isletalk/message/api/grpc"
	identityv1 "isletalk/message/api/identity/v1"
	messagev1 "isletalk/message/api/message/v1"
	"isletalk/message/application/service"
	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
)

// 契约测试（T025）：基于 message.proto 的 4 RPC 全覆盖。

type memMessages struct{ m map[string]entity.Message }

func (r *memMessages) Save(_ context.Context, msg entity.Message) (entity.Message, error) {
	r.m[msg.MessageID().String()] = msg
	return msg, nil
}
func (r *memMessages) FindByID(_ context.Context, id valueobject.MessageId) (entity.Message, error) {
	if m, ok := r.m[id.String()]; ok {
		return m, nil
	}
	return entity.Message{}, port.ErrNotFound
}
func (r *memMessages) ListByConversation(_ context.Context, convID string, beforeSeq uint64, limit int) ([]entity.Message, error) {
	var all []entity.Message
	for _, m := range r.m {
		if m.ConversationID() == convID && (beforeSeq == 0 || m.Seq() < beforeSeq) {
			all = append(all, m)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Seq() > all[j].Seq() })
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
func (r *memMessages) MarkRead(_ context.Context, _ string, _ valueobject.UserId, _ uint64) (int64, error) {
	return 0, nil
}
func (r *memMessages) Recall(_ context.Context, id valueobject.MessageId) error {
	if m, ok := r.m[id.String()]; ok {
		m.Recall()
		r.m[id.String()] = m
	}
	return nil
}

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
func (r *memConvs) ListByUser(_ context.Context, _ valueobject.UserId, _ int) ([]entity.Conversation, error) {
	return nil, nil
}
func (r *memConvs) NextSeq(_ context.Context, _ string) (uint64, error) { return 1, nil }
func (r *memConvs) TouchOnSend(_ context.Context, _, _ string) error    { return nil }
func (r *memConvs) ResetUnread(_ context.Context, _ string) error       { return nil }

func (r *memConvs) TouchOnRecall(_ context.Context, _, _ string) error { return nil }

type memRel struct{ allowed, blocked bool }

func (r memRel) CanMessage(_ context.Context, _, _ valueobject.UserId) (bool, bool, error) {
	return r.allowed, r.blocked, nil
}

func uid(s string) *identityv1.UserId { return &identityv1.UserId{Value: s} }

func newServer(allowed bool) *grpcserver.MessageGRPCServer {
	msgs := &memMessages{m: map[string]entity.Message{}}
	convs := &memConvs{m: map[string]entity.Conversation{}}
	return grpcserver.NewMessageGRPCServer(
		service.NewSendMessageService(msgs, convs, memRel{allowed: allowed}, nil, nil),
		service.NewListMessagesService(msgs),
		service.NewRecallService(msgs, convs, nil),
		service.NewMarkReadService(msgs, convs, nil, nil),
	)
}

func TestContractSendMessage(t *testing.T) {
	srv := newServer(true)
	resp, err := srv.SendMessage(context.Background(), &messagev1.SendMessageRequest{
		MessageId: "m1", ConversationId: "direct:a:b", SenderId: uid("a"),
		Type: messagev1.MessageType_MESSAGE_TYPE_TEXT, Content: "hi",
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if resp.Message.GetStatus() != messagev1.MessageStatus_MESSAGE_STATUS_DELIVERED {
		t.Fatalf("status = %v, want DELIVERED", resp.Message.GetStatus())
	}
}

func TestContractSendMessageNotFriends(t *testing.T) {
	srv := newServer(false)
	_, err := srv.SendMessage(context.Background(), &messagev1.SendMessageRequest{
		MessageId: "m1", ConversationId: "direct:a:b", SenderId: uid("a"),
		Type: messagev1.MessageType_MESSAGE_TYPE_TEXT, Content: "hi",
	})
	if err == nil {
		t.Fatal("non-friends should be rejected")
	}
}

func TestContractListMessages(t *testing.T) {
	srv := newServer(true)
	if _, err := srv.SendMessage(context.Background(), &messagev1.SendMessageRequest{
		MessageId: "m1", ConversationId: "direct:a:b", SenderId: uid("a"),
		Type: messagev1.MessageType_MESSAGE_TYPE_TEXT, Content: "hi",
	}); err != nil {
		t.Fatal(err)
	}
	resp, err := srv.ListMessages(context.Background(), &messagev1.ListMessagesRequest{
		ConversationId: "direct:a:b", PageSize: 50,
	})
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(resp.Messages))
	}
}

func TestContractRecallMessage(t *testing.T) {
	srv := newServer(true)
	if _, err := srv.SendMessage(context.Background(), &messagev1.SendMessageRequest{
		MessageId: "m1", ConversationId: "direct:a:b", SenderId: uid("a"),
		Type: messagev1.MessageType_MESSAGE_TYPE_TEXT, Content: "hi",
	}); err != nil {
		t.Fatal(err)
	}
	resp, err := srv.RecallMessage(context.Background(), &messagev1.RecallMessageRequest{
		MessageId: "m1", SenderId: uid("a"),
	})
	if err != nil {
		t.Fatalf("RecallMessage: %v", err)
	}
	if !resp.GetRecalled() {
		t.Fatal("recall should succeed")
	}
}

func TestContractMarkRead(t *testing.T) {
	srv := newServer(true)
	if _, err := srv.SendMessage(context.Background(), &messagev1.SendMessageRequest{
		MessageId: "m1", ConversationId: "direct:a:b", SenderId: uid("a"),
		Type: messagev1.MessageType_MESSAGE_TYPE_TEXT, Content: "hi",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := srv.MarkRead(context.Background(), &messagev1.MarkReadRequest{
		ConversationId: "direct:a:b", UserId: uid("b"), LastReadSeq: 1,
	}); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
}
