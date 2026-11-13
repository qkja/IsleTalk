package service

import (
	"context"
	"errors"
	"testing"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// —— 内存 mock ——

type memFriendships struct{ m map[string]entity.Friendship }

func newMemFriendships() *memFriendships  { return &memFriendships{m: map[string]entity.Friendship{}} }
func key2(a, b valueobject.UserId) string { return a.String() + ":" + b.String() }

func (r *memFriendships) Create(_ context.Context, f entity.Friendship) (entity.Friendship, error) {
	r.m[key2(f.UserID(), f.FriendID())] = f
	return f, nil
}
func (r *memFriendships) Find(_ context.Context, a, b valueobject.UserId) (entity.Friendship, error) {
	if f, ok := r.m[key2(a, b)]; ok {
		return f, nil
	}
	return entity.Friendship{}, port.ErrNotFound
}
func (r *memFriendships) ListByUser(_ context.Context, uid valueobject.UserId) ([]entity.Friendship, error) {
	out := []entity.Friendship{}
	for _, f := range r.m {
		if f.UserID() == uid {
			out = append(out, f)
		}
	}
	return out, nil
}
func (r *memFriendships) Delete(_ context.Context, a, b valueobject.UserId) error {
	if f, ok := r.m[key2(a, b)]; ok {
		f.Delete()
		r.m[key2(a, b)] = f
	}
	return nil
}

type memRequests struct {
	m map[string]entity.FriendRequest
}

func newMemRequests() *memRequests { return &memRequests{m: map[string]entity.FriendRequest{}} }
func (r *memRequests) Create(_ context.Context, req entity.FriendRequest) (entity.FriendRequest, error) {
	r.m[key2(req.InitiatorID(), req.TargetID())] = req
	return req, nil
}
func (r *memRequests) Find(_ context.Context, a, b valueobject.UserId) (entity.FriendRequest, error) {
	if req, ok := r.m[key2(a, b)]; ok {
		return req, nil
	}
	return entity.FriendRequest{}, port.ErrNotFound
}
func (r *memRequests) ListPendingByTarget(_ context.Context, tid valueobject.UserId) ([]entity.FriendRequest, error) {
	out := []entity.FriendRequest{}
	for _, req := range r.m {
		if req.TargetID() == tid && req.IsPending() {
			out = append(out, req)
		}
	}
	return out, nil
}
func (r *memRequests) ListHistoryByTarget(_ context.Context, tid valueobject.UserId) ([]entity.FriendRequest, error) {
	out := []entity.FriendRequest{}
	for _, req := range r.m {
		if req.TargetID() == tid && !req.IsPending() {
			out = append(out, req)
		}
	}
	return out, nil
}
func (r *memRequests) Update(_ context.Context, req entity.FriendRequest) error {
	r.m[key2(req.InitiatorID(), req.TargetID())] = req
	return nil
}

type memBlocks struct{ m map[string]entity.Block }

func newMemBlocks() *memBlocks { return &memBlocks{m: map[string]entity.Block{}} }
func (r *memBlocks) Set(_ context.Context, owner, target valueobject.UserId, blocked bool) error {
	if blocked {
		r.m[key2(owner, target)] = entity.NewBlock(owner, target)
	} else {
		if b, ok := r.m[key2(owner, target)]; ok {
			b.Unblock()
			r.m[key2(owner, target)] = b
		}
	}
	return nil
}
func (r *memBlocks) Find(_ context.Context, owner, target valueobject.UserId) (entity.Block, error) {
	if b, ok := r.m[key2(owner, target)]; ok {
		return b, nil
	}
	return entity.Block{}, port.ErrNotFound
}

type memIdentity struct{ allow map[string]bool }

func newMemIdentity() *memIdentity { return &memIdentity{allow: map[string]bool{}} }
func (m *memIdentity) GetUser(_ context.Context, uid valueobject.UserId) (port.IdentityUser, error) {
	allow, ok := m.allow[uid.String()]
	if !ok {
		return port.IdentityUser{}, errors.New("user not found")
	}
	return port.IdentityUser{ID: uid, PublicName: "user-" + uid.String(), AllowStrangerAdd: allow}, nil
}

type memRemarks struct{ m map[string]entity.Remark }

func newMemRemarks() *memRemarks { return &memRemarks{m: map[string]entity.Remark{}} }
func (r *memRemarks) Save(_ context.Context, remark entity.Remark) (entity.Remark, error) {
	r.m[key2(remark.OwnerID(), remark.FriendID())] = remark
	return remark, nil
}
func (r *memRemarks) Find(_ context.Context, owner, friend valueobject.UserId) (entity.Remark, error) {
	if remark, ok := r.m[key2(owner, friend)]; ok {
		return remark, nil
	}
	return entity.Remark{}, port.ErrNotFound
}

func mustUID(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
