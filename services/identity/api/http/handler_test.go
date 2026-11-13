package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"isletalk/identity/application/service"
	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
	"isletalk/identity/infrastructure/auth"
	"isletalk/identity/infrastructure/cache"
)

// —— 内存 mock ——

type memUsers struct{ m map[string]entity.User }

func (m *memUsers) Create(_ context.Context, u entity.User) (entity.User, error) { return u, nil }
func (m *memUsers) FindByID(_ context.Context, id valueobject.UserId) (entity.User, error) {
	if u, ok := m.m[id.String()]; ok {
		return u, nil
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) FindByPhoneHash(_ context.Context, h valueobject.PhoneHash) (entity.User, error) {
	for _, u := range m.m {
		if u.PhoneHash() == h {
			return u, nil
		}
	}
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) FindByPublicName(_ context.Context, n valueobject.PublicName) (entity.User, error) {
	return entity.User{}, port.ErrNotFound
}
func (m *memUsers) Update(_ context.Context, u entity.User) error { return nil }

type memSessions struct{ m map[string]entity.Session }

func (m *memSessions) Create(_ context.Context, s entity.Session) (entity.Session, error) {
	return s, nil
}
func (m *memSessions) FindByID(_ context.Context, id string) (entity.Session, error) {
	return entity.Session{}, port.ErrNotFound
}
func (m *memSessions) ListByUser(_ context.Context, _ valueobject.UserId) ([]entity.Session, error) {
	return nil, nil
}
func (m *memSessions) Revoke(_ context.Context, _ string) error                      { return nil }
func (m *memSessions) RevokeAllByUser(_ context.Context, _ valueobject.UserId) error { return nil }
func (m *memSessions) Touch(_ context.Context, _ string) error                       { return nil }

type memCodes struct{ m map[string]string }

func (m *memCodes) Save(_ context.Context, p valueobject.PhoneHash, code string) error {
	m.m[p.String()] = code
	return nil
}
func (m *memCodes) Verify(_ context.Context, p valueobject.PhoneHash, code string) (bool, error) {
	return m.m[p.String()] == code, nil
}

type memSender struct{ sent string }

func (m *memSender) Send(_ context.Context, _ valueobject.PhoneHash, code string) error {
	m.sent = code
	return nil
}

func newTestHandler() *Handler {
	users := &memUsers{m: map[string]entity.User{}}
	sessions := &memSessions{m: map[string]entity.Session{}}
	codes := &memCodes{m: map[string]string{}}
	sender := &memSender{}
	tokens := auth.NewTokenIssuer()
	passwords := auth.NewPasswordHasher()
	phones := auth.NewPhoneHasher()
	ids := auth.NewIDGenerator()

	register := service.NewRegisterService(users, sessions, tokens, codes, phones, passwords, ids)
	login := service.NewLoginService(users, sessions, tokens, codes, phones, passwords, ids)
	reset := service.NewResetPasswordService(users, sessions, codes, phones, passwords)
	sessionSvc := service.NewSessionService(sessions)

	// limiter 传 nil：测试不触发限流（rateLimit 内 nil 检查跳过）
	return NewHandler(register, login, reset, sessionSvc, codes, sender, phones, nil)
}

func doRequest(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)
	return rec
}

func decodeEnv(t *testing.T, rec *httptest.ResponseRecorder) Envelope {
	t.Helper()
	var env Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return env
}

func TestLoginInvalidCredentialAntiEnumeration(t *testing.T) {
	h := newTestHandler()

	// 未注册手机号登录
	rec := doRequest(t, h, "POST", "/api/v1/login", `{"phone":"13800000000","password":"x","deviceId":"d1"}`)
	env1 := decodeEnv(t, rec)
	if rec.Code != http.StatusUnauthorized || env1.Success {
		t.Fatalf("login unregistered: code=%d success=%v", rec.Code, env1.Success)
	}
	if env1.Meta == nil || env1.Meta.Code != "AUTH_UNAUTHENTICATED" {
		t.Fatalf("login unregistered code=%v", env1.Meta)
	}

	// 错误密码（同一 code，防枚举 SC-003）
	rec2 := doRequest(t, h, "POST", "/api/v1/login", `{"phone":"13800000000","password":"y","deviceId":"d1"}`)
	env2 := decodeEnv(t, rec2)
	if env2.Meta == nil || env2.Meta.Code != env1.Meta.Code {
		t.Fatalf("anti-enumeration violated: %v vs %v", env1.Meta, env2.Meta)
	}
}

func TestRegisterInvalidCode(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/v1/register",
		`{"phone":"13800000000","code":"wrong","password":"secret","publicName":"alice","deviceId":"d1"}`)
	env := decodeEnv(t, rec)
	if env.Success || env.Meta == nil || env.Meta.Code != "RATE_VERIFY_CODE_COOLDOWN" {
		t.Fatalf("register invalid code: success=%v meta=%v", env.Success, env.Meta)
	}
}

func TestSendCodeEnvelope(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/v1/verify-code", `{"phone":"13800000000"}`)
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("send code: code=%d success=%v", rec.Code, env.Success)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("sensitive response should have Cache-Control: no-store")
	}
}

func TestResetPasswordInvalidCode(t *testing.T) {
	h := newTestHandler()

	rec := doRequest(t, h, "POST", "/api/v1/reset-password",
		`{"phone":"13800000000","code":"wrong","newPassword":"secret"}`)
	env := decodeEnv(t, rec)
	if env.Success {
		t.Fatal("reset with wrong code should fail")
	}
}

// failingCodes Save 返回错误，用于测错误路径。
type failingCodes struct{ err error }

func (f failingCodes) Save(context.Context, valueobject.PhoneHash, string) error { return f.err }
func (f failingCodes) Verify(context.Context, valueobject.PhoneHash, string) (bool, error) {
	return false, f.err
}

func TestSendCodeSaveFails(t *testing.T) {
	h := newTestHandler()
	h.codes = failingCodes{err: errBoom}

	rec := doRequest(t, h, "POST", "/api/v1/verify-code", `{"phone":"13800000000"}`)
	env := decodeEnv(t, rec)
	if env.Success {
		t.Fatal("send code should fail when save fails")
	}
}

func TestRateLimitExceeded(t *testing.T) {
	client := redisClientForTest(t)
	if client == nil {
		t.Skip("redis not available")
	}
	limiter := cache.NewRateLimiter(client, time.Minute, 1) // limit 1：第二次超限
	h := newTestHandler()
	h.limiter = limiter

	// 第一次不超限（limit=1），第二次超限
	doRequest(t, h, "POST", "/api/v1/login", `{"phone":"13800000000","password":"x","deviceId":"d1"}`)
	rec := doRequest(t, h, "POST", "/api/v1/login", `{"phone":"13800000000","password":"x","deviceId":"d1"}`)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("code = %d, want 429", rec.Code)
	}
}

var errBoom = errors.New("boom")

func redisClientForTest(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type failingSender struct{ err error }

func (f failingSender) Send(context.Context, valueobject.PhoneHash, string) error { return f.err }

func TestSendCodeSendFails(t *testing.T) {
	h := newTestHandler()
	h.sender = failingSender{err: errBoom}

	rec := doRequest(t, h, "POST", "/api/v1/verify-code", `{"phone":"13800000000"}`)
	env := decodeEnv(t, rec)
	if env.Success {
		t.Fatal("send code should fail when send fails")
	}
}

func TestRegisterBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/register", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestLoginBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/login", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestResetPasswordBodyInvalid(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/reset-password", `{`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	r1 := httptest.NewRequest("GET", "/", nil)
	r1.RemoteAddr = "1.2.3.4:5678"
	if got := clientIP(r1); got != "1.2.3.4" {
		t.Fatalf("clientIP = %q, want 1.2.3.4", got)
	}
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.RemoteAddr = "no-port"
	if got := clientIP(r2); got != "no-port" {
		t.Fatalf("clientIP noport = %q, want no-port", got)
	}
}

func TestGenerateCode(t *testing.T) {
	for range 100 {
		c := generateCode()
		if len(c) != 6 {
			t.Fatalf("generateCode length = %d, want 6", len(c))
		}
		for _, ch := range c {
			if ch < '0' || ch > '9' {
				t.Fatalf("generateCode non-digit: %q", c)
			}
		}
	}
}

func TestLoginByCodeInvalidEnvelope(t *testing.T) {
	h := newTestHandler()

	// code 非空走验证码登录，验证码错误统一 401（防枚举）
	rec := doRequest(t, h, "POST", "/api/v1/login", `{"phone":"13800000000","code":"wrong","deviceId":"d1"}`)
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusUnauthorized || env.Success {
		t.Fatalf("login by code: code=%d success=%v", rec.Code, env.Success)
	}
}

func TestLogout(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/logout", `{"sessionId":"s1"}`)
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("logout: code=%d success=%v", rec.Code, env.Success)
	}
}

func TestLogoutMissingSession(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "POST", "/api/v1/logout", `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("logout missing session: code = %d, want 400", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/healthz", "")
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("healthz: code=%d success=%v", rec.Code, env.Success)
	}
}

func TestReadyz(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/readyz", "")
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("readyz: code=%d success=%v", rec.Code, env.Success)
	}
}

func TestStartupz(t *testing.T) {
	h := newTestHandler()
	rec := doRequest(t, h, "GET", "/startupz", "")
	env := decodeEnv(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("startupz: code=%d success=%v", rec.Code, env.Success)
	}
}

type failingSessions struct{ err error }

func (f failingSessions) Create(context.Context, entity.Session) (entity.Session, error) {
	return entity.Session{}, f.err
}
func (f failingSessions) FindByID(context.Context, string) (entity.Session, error) {
	return entity.Session{}, port.ErrNotFound
}
func (f failingSessions) ListByUser(context.Context, valueobject.UserId) ([]entity.Session, error) {
	return nil, nil
}
func (f failingSessions) Revoke(context.Context, string) error                      { return f.err }
func (f failingSessions) RevokeAllByUser(context.Context, valueobject.UserId) error { return f.err }
func (f failingSessions) Touch(context.Context, string) error                       { return f.err }

func TestLogoutRevokeFails(t *testing.T) {
	h := newTestHandler()
	h.sessions = service.NewSessionService(failingSessions{err: errBoom})

	rec := doRequest(t, h, "POST", "/api/v1/logout", `{"sessionId":"s1"}`)
	env := decodeEnv(t, rec)
	if env.Success {
		t.Fatal("logout should fail when revoke fails")
	}
}

func TestErrorEnvelopeLocaleHeader(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBufferString(`{"phone":"13800000000","password":"x","deviceId":"d1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	rec := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	env := decodeEnv(t, rec)
	if env.Message == nil || *env.Message != "Incorrect account or password" {
		t.Fatalf("en message = %v", env.Message)
	}
}
