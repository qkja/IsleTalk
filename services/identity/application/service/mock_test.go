package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// —— 内存 mock，仅用于测试（无 IO，宪法原则 V）——

type memUserRepo struct {
	mu    sync.Mutex
	users map[string]entity.User // key: phone hash
}

func newMemUserRepo() *memUserRepo { return &memUserRepo{users: map[string]entity.User{}} }

func (m *memUserRepo) Create(_ context.Context, u entity.User) (entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.PhoneHash().String()] = u
	return u, nil
}
func (m *memUserRepo) FindByID(_ context.Context, id valueobject.UserId) (entity.User, error) {
	for _, u := range m.users {
		if u.ID() == id {
			return u, nil
		}
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUserRepo) FindByPhoneHash(_ context.Context, h valueobject.PhoneHash) (entity.User, error) {
	if u, ok := m.users[h.String()]; ok {
		return u, nil
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUserRepo) FindByPublicName(_ context.Context, n valueobject.PublicName) (entity.User, error) {
	for _, u := range m.users {
		if u.PublicName().String() == n.String() {
			return u, nil
		}
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUserRepo) Update(_ context.Context, u entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[u.PhoneHash().String()] = u
	return nil
}

type memSessionRepo struct {
	mu       sync.Mutex
	sessions map[string]entity.Session
	revoked  int
}

func newMemSessionRepo() *memSessionRepo {
	return &memSessionRepo{sessions: map[string]entity.Session{}}
}

func (m *memSessionRepo) Create(_ context.Context, s entity.Session) (entity.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID()] = s
	return s, nil
}
func (m *memSessionRepo) FindByID(_ context.Context, id string) (entity.Session, error) {
	if s, ok := m.sessions[id]; ok {
		return s, nil
	}
	return entity.Session{}, port.ErrNotFound
}
func (m *memSessionRepo) ListByUser(_ context.Context, uid valueobject.UserId) ([]entity.Session, error) {
	out := []entity.Session{}
	for _, s := range m.sessions {
		if s.UserID() == uid {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *memSessionRepo) Revoke(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.Revoke()
		m.sessions[id] = s
	}
	return nil
}
func (m *memSessionRepo) RevokeAllByUser(_ context.Context, uid valueobject.UserId) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, s := range m.sessions {
		if s.UserID() == uid {
			s.Revoke()
			m.sessions[k] = s
		}
	}
	return nil
}
func (m *memSessionRepo) Touch(_ context.Context, id string) error { return nil }

type stubTokens struct{ n int }

func (s *stubTokens) Issue(_ context.Context, uid valueobject.UserId) (port.TokenPair, error) {
	s.n++
	return port.TokenPair{
		AccessToken:  fmt.Sprintf("access-%d", s.n),
		RefreshToken: fmt.Sprintf("refresh-%d", s.n),
	}, nil
}
func (s *stubTokens) Verify(_ context.Context, _ string) (valueobject.UserId, error) {
	return valueobject.UserId{}, nil
}
func (s *stubTokens) Revoke(_ context.Context, _ valueobject.UserId) error { return nil }

type memCodes struct {
	codes map[string]string
}

func newMemCodes() *memCodes { return &memCodes{codes: map[string]string{}} }
func (m *memCodes) Save(_ context.Context, p valueobject.PhoneHash, code string) error {
	m.codes[p.String()] = code
	return nil
}
func (m *memCodes) Verify(_ context.Context, p valueobject.PhoneHash, code string) (bool, error) {
	if c, ok := m.codes[p.String()]; ok && c == code {
		delete(m.codes, p.String()) // 一次性
		return true, nil
	}
	return false, nil
}

type plainHasher struct{}

func (plainHasher) Hash(plain string) (valueobject.PasswordHash, error) {
	h, _ := valueobject.NewPasswordHash("hash:" + plain)
	return h, nil
}
func (plainHasher) Verify(plain string, hash valueobject.PasswordHash) bool {
	return hash.String() == "hash:"+plain
}

type plainPhone struct{}

func (plainPhone) Hash(phone string) valueobject.PhoneHash {
	h, _ := valueobject.NewPhoneHash("phone:" + phone)
	return h
}

type seqID struct{ n int }

func (s *seqID) NewID() valueobject.UserId {
	s.n++
	v, _ := valueobject.NewUserId(fmt.Sprintf("id-%d", s.n))
	return v
}

func mustPhone(t *testing.T, s string) valueobject.PhoneHash {
	t.Helper()
	h, err := valueobject.NewPhoneHash("phone:" + s)
	if err != nil {
		t.Fatalf("NewPhoneHash: %v", err)
	}
	return h
}

func idVal(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatalf("NewUserId: %v", err)
	}
	return v
}

func mustName(t *testing.T, s string) valueobject.PublicName {
	t.Helper()
	v, err := valueobject.NewPublicName(s)
	if err != nil {
		t.Fatalf("NewPublicName: %v", err)
	}
	return v
}

func now() time.Time { return time.Now().UTC() }

type memVerificationRepo struct {
	mu   sync.Mutex
	data map[string]entity.Verification
}

func newMemVerificationRepo() *memVerificationRepo {
	return &memVerificationRepo{data: map[string]entity.Verification{}}
}

func (m *memVerificationRepo) Save(_ context.Context, v entity.Verification) (entity.Verification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[v.UserID().String()] = v
	return v, nil
}
func (m *memVerificationRepo) FindByUser(_ context.Context, uid valueobject.UserId) (entity.Verification, error) {
	if v, ok := m.data[uid.String()]; ok {
		return v, nil
	}
	return entity.Verification{}, port.ErrNotFound
}
func (m *memVerificationRepo) Revoke(_ context.Context, uid valueobject.UserId) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.data[uid.String()]; ok {
		v.Revoke()
		m.data[uid.String()] = v
	}
	return nil
}
func (m *memVerificationRepo) DeleteExpired(_ context.Context, before time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for k, v := range m.data {
		if v.RetentionUntil().Before(before) {
			delete(m.data, k)
			n++
		}
	}
	return n, nil
}

// newUserForTest 创建一个已设密码的预置用户。
func newUserForTest(t *testing.T, phone, password, name string) (entity.User, error) {
	t.Helper()
	id, _ := valueobject.NewUserId("test-id")
	pn, err := valueobject.NewPublicName(name)
	if err != nil {
		t.Fatalf("NewPublicName: %v", err)
	}
	ph := mustPhone(t, phone)
	u, err := entity.NewUser(id, pn, ph, "zh-CN")
	if err != nil {
		return entity.User{}, err
	}
	phash, _ := valueobject.NewPasswordHash("hash:" + password)
	u.SetPassword(phash)
	return u, nil
}

// failingUserRepo 所有方法返回错误，用于测错误路径。
type failingUserRepo struct{ err error }

func (f failingUserRepo) Create(context.Context, entity.User) (entity.User, error) {
	return entity.User{}, f.err
}
func (f failingUserRepo) FindByID(context.Context, valueobject.UserId) (entity.User, error) {
	return entity.User{}, f.err
}
func (f failingUserRepo) FindByPhoneHash(context.Context, valueobject.PhoneHash) (entity.User, error) {
	return entity.User{}, f.err
}
func (f failingUserRepo) FindByPublicName(context.Context, valueobject.PublicName) (entity.User, error) {
	return entity.User{}, f.err
}
func (f failingUserRepo) Update(context.Context, entity.User) error { return f.err }

// failingSessionRepo 所有方法返回错误。
type failingSessionRepo struct{ err error }

func (f failingSessionRepo) Create(context.Context, entity.Session) (entity.Session, error) {
	return entity.Session{}, f.err
}
func (f failingSessionRepo) FindByID(context.Context, string) (entity.Session, error) {
	return entity.Session{}, f.err
}
func (f failingSessionRepo) ListByUser(context.Context, valueobject.UserId) ([]entity.Session, error) {
	return nil, f.err
}
func (f failingSessionRepo) Revoke(context.Context, string) error                      { return f.err }
func (f failingSessionRepo) RevokeAllByUser(context.Context, valueobject.UserId) error { return f.err }
func (f failingSessionRepo) Touch(context.Context, string) error                       { return f.err }

// failingTokens Issue 返回错误。
type failingTokens struct{ err error }

func (f failingTokens) Issue(context.Context, valueobject.UserId) (port.TokenPair, error) {
	return port.TokenPair{}, f.err
}
func (f failingTokens) Verify(context.Context, string) (valueobject.UserId, error) {
	return valueobject.UserId{}, f.err
}
func (f failingTokens) Revoke(context.Context, valueobject.UserId) error { return f.err }

// failingCreateRepo：FindByPhoneHash 返回 NotFound（让注册继续），Create 返回错误。
type failingCreateRepo struct{ err error }

func (f failingCreateRepo) Create(context.Context, entity.User) (entity.User, error) {
	return entity.User{}, f.err
}
func (f failingCreateRepo) FindByID(context.Context, valueobject.UserId) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (f failingCreateRepo) FindByPhoneHash(context.Context, valueobject.PhoneHash) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (f failingCreateRepo) FindByPublicName(context.Context, valueobject.PublicName) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (f failingCreateRepo) Update(context.Context, entity.User) error { return f.err }

// failingVerificationRepo 所有方法返回错误。
type failingVerificationRepo struct{ err error }

func (f failingVerificationRepo) Save(context.Context, entity.Verification) (entity.Verification, error) {
	return entity.Verification{}, f.err
}
func (f failingVerificationRepo) FindByUser(context.Context, valueobject.UserId) (entity.Verification, error) {
	return entity.Verification{}, f.err
}
func (f failingVerificationRepo) Revoke(context.Context, valueobject.UserId) error { return f.err }
func (f failingVerificationRepo) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, f.err
}
