package service

import (
	"context"
	"errors"
	"testing"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

func mustName(t *testing.T, s string) valueobject.PublicName {
	t.Helper()
	v, err := valueobject.NewPublicName(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

type fakeClaims struct {
	available bool
	err       error
}

func (f fakeClaims) Claim(context.Context, valueobject.PublicName, valueobject.UserId) error {
	return nil
}
func (f fakeClaims) Release(context.Context, valueobject.PublicName) error { return nil }
func (f fakeClaims) IsAvailable(context.Context, valueobject.PublicName) (bool, error) {
	return f.available, f.err
}

type fakeUsers struct {
	hasUser bool
	err     error
}

func (f fakeUsers) Create(context.Context, entity.User) (entity.User, error) {
	return entity.User{}, nil
}
func (f fakeUsers) FindByID(context.Context, valueobject.UserId) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (f fakeUsers) FindByPhoneHash(context.Context, valueobject.PhoneHash) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (f fakeUsers) FindByPublicName(context.Context, valueobject.PublicName) (entity.User, error) {
	if f.hasUser {
		return entity.User{}, nil
	}
	if f.err != nil {
		return entity.User{}, f.err
	}
	return entity.User{}, port.ErrNotFound
}
func (f fakeUsers) Update(context.Context, entity.User) error { return nil }

func TestNameUniquenessIsAvailable(t *testing.T) {
	ctx := context.Background()
	n := mustName(t, "alice")

	// 活跃用户持有同名 → 不可占用
	s := NewNameUniqueness(fakeClaims{available: true}, fakeUsers{hasUser: true})
	if ok, err := s.IsAvailable(ctx, n); err != nil || ok {
		t.Fatalf("held name: ok=%v err=%v, want false nil", ok, err)
	}

	// 仓储返回非 NotFound 错误 → 传播
	s = NewNameUniqueness(fakeClaims{}, fakeUsers{err: errors.New("db down")})
	if _, err := s.IsAvailable(ctx, n); err == nil {
		t.Fatal("non-notfound error should propagate")
	}

	// 无持有 → 委托 claims 判定为可用
	s = NewNameUniqueness(fakeClaims{available: true}, fakeUsers{hasUser: false})
	if ok, err := s.IsAvailable(ctx, n); err != nil || !ok {
		t.Fatalf("available: ok=%v err=%v, want true nil", ok, err)
	}

	// claims 判定为不可用（保留期占用）
	s = NewNameUniqueness(fakeClaims{available: false}, fakeUsers{hasUser: false})
	if ok, _ := s.IsAvailable(ctx, n); ok {
		t.Fatal("claims unavailable should be false")
	}
}
