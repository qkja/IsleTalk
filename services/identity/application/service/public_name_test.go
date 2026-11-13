package service

import (
	"context"
	"sync"
	"testing"

	"isletalk/identity/domain/service"
	"isletalk/identity/domain/valueobject"
)

// memNameClaimRepo 是账号名占用的内存 mock。
type memNameClaimRepo struct {
	mu     sync.Mutex
	claims map[string]bool // true = 占用或保留期
}

func newMemNameClaimRepo() *memNameClaimRepo {
	return &memNameClaimRepo{claims: map[string]bool{}}
}

func (m *memNameClaimRepo) Claim(_ context.Context, n valueobject.PublicName, _ valueobject.UserId) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[n.String()] = true
	return nil
}
func (m *memNameClaimRepo) Release(_ context.Context, n valueobject.PublicName) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claims[n.String()] = true // 保留期内仍占位
	return nil
}
func (m *memNameClaimRepo) IsAvailable(_ context.Context, n valueobject.PublicName) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.claims[n.String()], nil
}

func TestNameUniquenessIsAvailable(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	claims := newMemNameClaimRepo()
	svc := service.NewNameUniqueness(claims, users)

	// 空库：可占用
	ok, err := svc.IsAvailable(ctx, mustName(t, "alice"))
	if err != nil || !ok {
		t.Fatalf("IsAvailable(empty) = %v, %v, want true", ok, err)
	}

	// 占用后：不可用
	if err := claims.Claim(ctx, mustName(t, "alice"), idVal(t, "u1")); err != nil {
		t.Fatal(err)
	}
	ok, _ = svc.IsAvailable(ctx, mustName(t, "alice"))
	if ok {
		t.Fatal("claimed name should not be available")
	}
}

func TestPublicNameChange(t *testing.T) {
	ctx := context.Background()
	users := newMemUserRepo()
	claims := newMemNameClaimRepo()
	svc := NewPublicNameService(users, claims)

	u, _ := newUserForTest(t, "138", "secret", "alice")
	users.users[u.PhoneHash().String()] = u
	claims.Claim(ctx, mustName(t, "alice"), u.ID())

	// 改名为未占用的新名
	changed, err := svc.Change(ctx, u.ID(), "bob")
	if err != nil {
		t.Fatalf("Change err: %v", err)
	}
	if changed.PublicName().String() != "bob" {
		t.Fatalf("public name = %q, want bob", changed.PublicName())
	}
	// 内部 ID 不变（FR-014）
	if changed.ID() != u.ID() {
		t.Fatal("internal ID must not change")
	}
}
