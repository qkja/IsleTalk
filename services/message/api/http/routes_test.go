package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	h, _ := newHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 通过 mux 走 healthz，证明路由已挂载（其余端点由 handler_test.go 直接覆盖）
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz via mux status = %d, want 200", w.Code)
	}
}
