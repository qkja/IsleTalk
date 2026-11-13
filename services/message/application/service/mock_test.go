package service

import (
	"context"
	"sync"

	"isletalk/message/domain/entity"
	"isletalk/message/domain/port"
	"isletalk/message/domain/valueobject"
)

// 测试用内存仓储与客户端（无 IO，覆盖应用层用例逻辑）。

// mockRelationship 是 RelationshipClient 的内存实现，可配置授权结果。
type mockRelationship struct {
	allowed bool
	blocked bool
	err     error
	calls   int
}

func (m *mockRelationship) CanMessage(_ context.Context, _, _ valueobject.UserId) (bool, bool, error) {
	m.calls++
	return m.allowed, m.blocked, m.err
}

// memMessageRepo 是 MessageRepository 的内存实现。
type memMessageRepo struct {
	mu   sync.Mutex
	byID map[string]entity.Message
}

func newMemMessageRepo() *memMessageRepo {
	return &memMessageRepo{byID: map[string]entity.Message{}}
}

func (r *memMessageRepo) Save(_ context.Context, m entity.Message) (entity.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[m.MessageID().String()] = m
	return m, nil
}

func (r *memMessageRepo) FindByID(_ context.Context, id valueobject.MessageId) (entity.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.byID[id.String()]; ok {
		return m, nil
	}
	return entity.Message{}, port.ErrNotFound
}

func (r *memMessageRepo) ListByConversation(_ context.Context, conversationID string, beforeSeq uint64, limit int) ([]entity.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 会话内按 seq 倒序（列表按 seq 排序，模拟索引）
	all := []entity.Message{}
	for _, m := range r.byID {
		if m.ConversationID() == conversationID {
			all = append(all, m)
		}
	}
	// 简单冒泡按 seq 降序
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].Seq() > all[i].Seq() {
				all[i], all[j] = all[j], all[i]
			}
		}
	}
	out := []entity.Message{}
	for _, m := range all {
		if beforeSeq > 0 && m.Seq() >= beforeSeq {
			continue
		}
		out = append(out, m)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (r *memMessageRepo) MarkRead(_ context.Context, conversationID string, readerID valueobject.UserId, upToSeq uint64) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var n int64
	for id, m := range r.byID {
		if m.ConversationID() == conversationID && m.Seq() <= upToSeq &&
			!m.IsSentBy(readerID) && m.Status() == entity.MessageDelivered {
			m.MarkRead()
			r.byID[id] = m
			n++
		}
	}
	return n, nil
}

func (r *memMessageRepo) Recall(_ context.Context, id valueobject.MessageId) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.byID[id.String()]
	if !ok {
		return port.ErrNotFound
	}
	m.Recall()
	r.byID[id.String()] = m
	return nil
}

// memConversationRepo 是 ConversationRepository 的内存实现。
type memConversationRepo struct {
	mu   sync.Mutex
	byID map[string]entity.Conversation
	seq  map[string]uint64
}

func newMemConversationRepo() *memConversationRepo {
	return &memConversationRepo{byID: map[string]entity.Conversation{}, seq: map[string]uint64{}}
}

func (r *memConversationRepo) Save(_ context.Context, c entity.Conversation) (entity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[c.ConversationID()] = c
	return c, nil
}

func (r *memConversationRepo) FindByID(_ context.Context, conversationID string) (entity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.byID[conversationID]; ok {
		return c, nil
	}
	return entity.Conversation{}, port.ErrNotFound
}

func (r *memConversationRepo) ListByUser(_ context.Context, userID valueobject.UserId, limit int) ([]entity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []entity.Conversation{}
	for _, c := range r.byID {
		if c.HasParticipant(userID) {
			out = append(out, c)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *memConversationRepo) NextSeq(_ context.Context, conversationID string) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq[conversationID]++
	return r.seq[conversationID], nil
}

func (r *memConversationRepo) TouchOnSend(_ context.Context, conversationID, preview string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.byID[conversationID]; ok {
		c.SetPreview(preview)
		c.IncrementUnread()
		c.Restore()
		r.byID[conversationID] = c
	}
	return nil
}

func (r *memConversationRepo) ResetUnread(_ context.Context, conversationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.byID[conversationID]; ok {
		c.ResetUnread()
		r.byID[conversationID] = c
	}
	return nil
}

func (r *memConversationRepo) TouchOnRecall(_ context.Context, conversationID, preview string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.byID[conversationID]; ok {
		c.SetPreview(preview)
		r.byID[conversationID] = c
	}
	return nil
}
