package grpc

import (
	"context"
	"testing"

	identityv1 "isletalk/identity/api/identity/v1"
	"isletalk/identity/application/service"
	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// 内存 mock（gRPC 包内，避免跨包依赖测试 mock）。

type memUsers struct{ m map[string]entity.User }

func (m *memUsers) Create(_ context.Context, u entity.User) (entity.User, error) { return u, nil }
func (m *memUsers) FindByID(_ context.Context, id valueobject.UserId) (entity.User, error) {
	if u, ok := m.m[id.String()]; ok {
		return u, nil
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) FindByPhoneHash(_ context.Context, h valueobject.PhoneHash) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) FindByPublicName(_ context.Context, n valueobject.PublicName) (entity.User, error) {
	for _, u := range m.m {
		if u.PublicName() == n {
			return u, nil
		}
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) Update(_ context.Context, u entity.User) error { return nil }

type memSessions struct{ m map[string]entity.Session }

func (m *memSessions) Create(_ context.Context, s entity.Session) (entity.Session, error) {
	return s, nil
}
func (m *memSessions) FindByID(_ context.Context, id string) (entity.Session, error) {
	if s, ok := m.m[id]; ok {
		return s, nil
	}
	return entity.Session{}, port.ErrNotFound
}
func (m *memSessions) ListByUser(_ context.Context, uid valueobject.UserId) ([]entity.Session, error) {
	out := []entity.Session{}
	for _, s := range m.m {
		if s.UserID() == uid {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *memSessions) Revoke(_ context.Context, id string) error {
	if s, ok := m.m[id]; ok {
		s.Revoke()
		m.m[id] = s
	}
	return nil
}
func (m *memSessions) RevokeAllByUser(_ context.Context, uid valueobject.UserId) error { return nil }
func (m *memSessions) Touch(_ context.Context, id string) error                        { return nil }

func newTestServer() (*IdentityGRPCServer, *memUsers, *memSessions) {
	users := &memUsers{m: map[string]entity.User{}}
	sessions := &memSessions{m: map[string]entity.Session{}}
	svc := service.NewSessionService(sessions)
	return NewIdentityGRPCServer(users, svc), users, sessions
}

func mustID(t *testing.T, s string) valueobject.UserId {
	t.Helper()
	v, err := valueobject.NewUserId(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestGetUser(t *testing.T) {
	srv, users, _ := newTestServer()
	uid := mustID(t, "u-1")
	name, _ := valueobject.NewPublicName("alice")
	phone, _ := valueobject.NewPhoneHash("phone:138")
	u, _ := entity.NewUser(uid, name, phone, "zh-CN")
	users.m["u-1"] = u

	resp, err := srv.GetUser(context.Background(), &identityv1.GetUserRequest{Id: &identityv1.UserId{Value: "u-1"}})
	if err != nil {
		t.Fatalf("GetUser err: %v", err)
	}
	if resp.User.GetPublicName().GetValue() != "alice" {
		t.Fatalf("public name = %q", resp.User.GetPublicName().GetValue())
	}
	// 响应 MUST NOT 含手机号（proto 无 phone 字段）
}

func TestSearchByPublicNameNotFound(t *testing.T) {
	srv, _, _ := newTestServer()
	resp, err := srv.SearchByPublicName(context.Background(), &identityv1.SearchByPublicNameRequest{PublicName: "nobody"})
	if err != nil {
		t.Fatalf("SearchByPublicName err: %v", err)
	}
	if resp.User != nil {
		t.Fatal("not found should return empty user")
	}
}

func TestCheckUser(t *testing.T) {
	srv, users, _ := newTestServer()
	uid := mustID(t, "u-1")
	name, _ := valueobject.NewPublicName("alice")
	phone, _ := valueobject.NewPhoneHash("phone:138")
	u, _ := entity.NewUser(uid, name, phone, "zh-CN")
	users.m["u-1"] = u

	resp, _ := srv.CheckUser(context.Background(), &identityv1.CheckUserRequest{Id: &identityv1.UserId{Value: "u-1"}})
	if !resp.Active {
		t.Fatal("active user should return true")
	}
	// 不存在 → false（不区分不存在/封禁）
	resp2, _ := srv.CheckUser(context.Background(), &identityv1.CheckUserRequest{Id: &identityv1.UserId{Value: "u-x"}})
	if resp2.Active {
		t.Fatal("nonexistent user should return false")
	}
}

func TestRevokeSession(t *testing.T) {
	srv, _, sessions := newTestServer()
	uid := mustID(t, "u-1")
	sessions.m["s-1"] = entity.NewSession("s-1", uid, "dev-1")

	_, err := srv.RevokeSession(context.Background(), &identityv1.RevokeSessionRequest{SessionId: "s-1"})
	if err != nil {
		t.Fatalf("RevokeSession err: %v", err)
	}
	if !sessions.m["s-1"].Revoked() {
		t.Fatal("session should be revoked")
	}
}
