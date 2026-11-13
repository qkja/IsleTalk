package http

import (
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"strings"

	"isletalk/identity/application/service"
	"isletalk/identity/domain/port"
	"isletalk/identity/infrastructure/auth"
	"isletalk/identity/infrastructure/cache"
)

// Handler 是客户端信封接口（T049）：登录/注册/找回密码/验证码下发。
// 走统一信封 {success,data,message,meta}（宪法原则 III），与 gRPC 服务间接口分离。
type Handler struct {
	register *service.RegisterService
	login    *service.LoginService
	reset    *service.ResetPasswordService
	sessions *service.SessionService
	codes    service.VerificationCodeStore
	sender   port.VerificationCodeSender
	phones   service.PhoneHasher
	limiter  *cache.RateLimiter
}

func NewHandler(
	register *service.RegisterService,
	login *service.LoginService,
	reset *service.ResetPasswordService,
	sessions *service.SessionService,
	codes service.VerificationCodeStore,
	sender port.VerificationCodeSender,
	phones service.PhoneHasher,
	limiter *cache.RateLimiter,
) *Handler {
	return &Handler{
		register: register,
		login:    login,
		reset:    reset,
		sessions: sessions,
		codes:    codes,
		sender:   sender,
		phones:   phones,
		limiter:  limiter,
	}
}

// RegisterRoutes 注册信封接口路由。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/verify-code", h.SendCode)
	mux.HandleFunc("POST /api/v1/register", h.Register)
	mux.HandleFunc("POST /api/v1/login", h.Login)
	mux.HandleFunc("POST /api/v1/reset-password", h.ResetPassword)
	mux.HandleFunc("POST /api/v1/logout", h.Logout)
	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /readyz", h.Readyz)
	mux.HandleFunc("GET /startupz", h.Startupz)
}

// SendCode 下发验证码（FR-001）。
func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	locale := localeFromRequest(r)
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_arg_invalid", locale)))
		return
	}
	if req.Phone == "" {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_phone_required", locale)))
		return
	}
	// 风控：账号 + IP 双维度限流（T045，宪法 E 组）
	if h.rateLimit(w, r, "phone:"+req.Phone) || h.rateLimit(w, r, "ip:"+clientIP(r)) {
		return
	}
	phoneHash := h.phones.Hash(req.Phone)
	code := generateCode()

	if err := h.codes.Save(r.Context(), phoneHash, code); err != nil {
		slog.Error("send code save failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err, locale))
		return
	}
	if err := h.sender.Send(r.Context(), phoneHash, code); err != nil {
		slog.Error("send code send failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err, locale))
		return
	}
	slog.Info("send code success", "phone_hash", phoneHash.String())
	writeJSON(w, http.StatusOK, ok(nil))
}

// Register 手机号注册（FR-001/002）。
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	locale := localeFromRequest(r)
	var req struct {
		Phone      string `json:"phone"`
		Code       string `json:"code"`
		Password   string `json:"password"`
		PublicName string `json:"publicName"`
		DeviceID   string `json:"deviceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_arg_invalid", locale)))
		return
	}
	// 风控：账号 + IP 双维度限流（T045）
	if h.rateLimit(w, r, "phone:"+req.Phone) || h.rateLimit(w, r, "ip:"+clientIP(r)) {
		return
	}

	res, err := h.register.Register(r.Context(), service.RegisterInput{
		Phone:      req.Phone,
		Code:       req.Code,
		Password:   req.Password,
		PublicName: req.PublicName,
		DeviceID:   req.DeviceID,
	})
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorEnvelope(err, locale))
		return
	}

	writeJSON(w, http.StatusOK, ok(map[string]interface{}{
		"userId":       res.User.ID().String(),
		"publicName":   res.User.PublicName().String(),
		"sessionId":    res.SessionID,
		"accessToken":  res.Tokens.AccessToken,
		"refreshToken": res.Tokens.RefreshToken,
	}))
}

// Login 登录（FR-003）：密码登录为主，验证码登录为辅（code 非空时走验证码）。
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	locale := localeFromRequest(r)
	var req struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Code     string `json:"code"`
		DeviceID string `json:"deviceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_arg_invalid", locale)))
		return
	}
	// 风控：账号 + IP 双维度限流（T045）
	if h.rateLimit(w, r, "phone:"+req.Phone) || h.rateLimit(w, r, "ip:"+clientIP(r)) {
		return
	}

	var res service.LoginResult
	var err error
	if req.Code != "" {
		res, err = h.login.LoginByCode(r.Context(), service.LoginByCodeInput{
			Phone:    req.Phone,
			Code:     req.Code,
			DeviceID: req.DeviceID,
		})
	} else {
		res, err = h.login.Login(r.Context(), service.LoginInput{
			Phone:    req.Phone,
			Password: req.Password,
			DeviceID: req.DeviceID,
		})
	}
	if err != nil {
		// 防枚举：未注册与密码错误统一 401 + 同 code（SC-003）
		writeJSON(w, http.StatusUnauthorized, errorEnvelope(err, locale))
		return
	}

	writeJSON(w, http.StatusOK, ok(map[string]interface{}{
		"userId":       res.User.ID().String(),
		"publicName":   res.User.PublicName().String(),
		"sessionId":    res.SessionID,
		"accessToken":  res.Tokens.AccessToken,
		"refreshToken": res.Tokens.RefreshToken,
	}))
}

// ResetPassword 找回密码（FR-004）。
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	locale := localeFromRequest(r)
	var req struct {
		Phone       string `json:"phone"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_arg_invalid", locale)))
		return
	}

	if err := h.reset.Reset(r.Context(), req.Phone, req.Code, req.NewPassword); err != nil {
		writeJSON(w, http.StatusUnauthorized, errorEnvelope(err, locale))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// Logout 退出登录（FR-005）：撤销当前会话。
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	locale := localeFromRequest(r)
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_arg_invalid", locale)))
		return
	}
	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", auth.Message("error_session_required", locale)))
		return
	}
	if err := h.sessions.Revoke(r.Context(), req.SessionID); err != nil {
		slog.Error("logout revoke failed", "err", err, "session_id", req.SessionID)
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err, locale))
		return
	}
	slog.Info("logout success", "session_id", req.SessionID)
	writeJSON(w, http.StatusOK, ok(nil))
}

// rateLimit 风控限流（T045，宪法 E 组：登录/验证码按账号 + IP 双维度限流）。
// 超限返回 true 表示已拒绝。
func (h *Handler) rateLimit(w http.ResponseWriter, r *http.Request, key string) bool {
	if h.limiter == nil {
		return false
	}
	ok, err := h.limiter.Allow(r.Context(), key)
	if err != nil {
		slog.Error("rate limit check failed", "err", err)
		return false
	}
	if !ok {
		writeJSON(w, http.StatusTooManyRequests, fail("RATE_LIMITED", auth.Message("error_rate_limited", localeFromRequest(r))))
		return true
	}
	return false
}

// localeFromRequest 从 Accept-Language 提取 locale，回退 zh-CN（FR-019）。
func localeFromRequest(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")
	if strings.HasPrefix(strings.ToLower(lang), "en") {
		return "en-US"
	}
	return "zh-CN"
}

// clientIP 从 RemoteAddr 提取客户端 IP。
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// generateCode 生成 6 位数字验证码。
func generateCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "000000"
	}
	return string(pad6(n.Int64()))
}

func pad6(n int64) []byte {
	b := []byte("000000")
	for i := 5; i >= 0 && n > 0; i-- {
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return b
}

// writeJSON 写统一信封响应（敏感响应带 Cache-Control: no-store）。
func writeJSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
