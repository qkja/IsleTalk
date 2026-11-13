package http

import (
	"context"
	"testing"

	"isletalk/relationship/application/service"
	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// —— 精简内存 mock，用于 handler 成功路径测试 ——

type memF struct{ m map[string]entity.Friendship }

func (r *memF) Create(_ context.Context, f entity.Friendship) (entity.Friendship, error) {
	r.m[f.UserID().String()+":"+f.FriendID().String()] = f
	return f, nil
}
func (r *memF) Find(_ context.Context, a, b valueobject.UserId) (entity.Friendship, error) {
	if f, ok := r.m[a.String()+":"+b.String()]; ok {
		return f, nil
	}
	return entity.Friendship{}, port.ErrNotFound
}
func (r *memF) ListByUser(context.Context, valueobject.UserId) ([]entity.Friendship, error) {
	return nil, nil
}
func (r *memF) Delete(_ context.Context, a, b valueobject.UserId) error { return nil }

type memR struct {
	m map[string]entity.FriendRequest
}

func (r *memR) Create(_ context.Context, req entity.FriendRequest) (entity.FriendRequest, error) {
	r.m[req.InitiatorID().String()+":"+req.TargetID().String()] = req
	return req, nil
}
func (r *memR) Find(_ context.Context, a, b valueobject.UserId) (entity.FriendRequest, error) {
	if req, ok := r.m[a.String()+":"+b.String()]; ok {
		return req, nil
	}
	return entity.FriendRequest{}, port.ErrNotFound
}
func (r *memR) ListPendingByTarget(context.Context, valueobject.UserId) ([]entity.FriendRequest, error) {
	return nil, nil
}
func (r *memR) ListHistoryByTarget(context.Context, valueobject.UserId) ([]entity.FriendRequest, error) {
	return nil, nil
}
func (r *memR) Update(_ context.Context, req entity.FriendRequest) error { return nil }

type memB struct{ m map[string]entity.Block }

func (r *memB) Set(_ context.Context, o, t valueobject.UserId, blocked bool) error {
	if blocked {
		r.m[o.String()+":"+t.String()] = entity.NewBlock(o, t)
	}
	return nil
}
func (r *memB) Find(_ context.Context, o, t valueobject.UserId) (entity.Block, error) {
	if b, ok := r.m[o.String()+":"+t.String()]; ok {
		return b, nil
	}
	return entity.Block{}, port.ErrNotFound
}

type memRk struct{ m map[string]entity.Remark }

func (r *memRk) Save(_ context.Context, remark entity.Remark) (entity.Remark, error) {
	r.m[remark.OwnerID().String()+":"+remark.FriendID().String()] = remark
	return remark, nil
}
func (r *memRk) Find(_ context.Context, o, f valueobject.UserId) (entity.Remark, error) {
	if remark, ok := r.m[o.String()+":"+f.String()]; ok {
		return remark, nil
	}
	return entity.Remark{}, port.ErrNotFound
}

type memI struct{}

func (memI) GetUser(_ context.Context, uid valueobject.UserId) (port.IdentityUser, error) {
	return port.IdentityUser{ID: uid, PublicName: "u", AllowStrangerAdd: true}, nil
}

func newSuccessHandler() *Handler {
	f := &memF{m: map[string]entity.Friendship{}}
	r := &memR{m: map[string]entity.FriendRequest{}}
	b := &memB{m: map[string]entity.Block{}}
	rk := &memRk{m: map[string]entity.Remark{}}
	return NewHandler(
		service.NewFriendRequestService(r, memI{}),
		service.NewRespondService(r, f),
		service.NewRemarkService(rk),
		service.NewBlockService(b),
		service.NewFriendListService(f),
	)
}

func TestSendRequestSuccess(t *testing.T) {
	h := newSuccessHandler()
	rec := doRequest(t, h, "POST", "/api/v1/friend-request", `{"initiatorId":"a","targetId":"b","message":"hi"}`)
	if rec.Code != 200 {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func TestRespondSuccess(t *testing.T) {
	h := newSuccessHandler()
	doRequest(t, h, "POST", "/api/v1/friend-request", `{"initiatorId":"a","targetId":"b","message":"hi"}`)
	rec := doRequest(t, h, "POST", "/api/v1/friend-respond", `{"initiatorId":"a","targetId":"b","accept":true}`)
	if rec.Code != 200 {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func TestSetRemarkSuccess(t *testing.T) {
	h := newSuccessHandler()
	rec := doRequest(t, h, "POST", "/api/v1/remark", `{"ownerId":"a","friendId":"b","content":"best"}`)
	if rec.Code != 200 {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func TestSetBlockSuccess(t *testing.T) {
	h := newSuccessHandler()
	rec := doRequest(t, h, "POST", "/api/v1/block", `{"ownerId":"a","targetId":"b","blocked":true}`)
	if rec.Code != 200 {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}

func TestDeleteFriendSuccess(t *testing.T) {
	h := newSuccessHandler()
	rec := doRequest(t, h, "DELETE", "/api/v1/friend", `{"userId":"a","friendId":"b"}`)
	if rec.Code != 200 {
		t.Fatalf("code = %d, want 200", rec.Code)
	}
}
