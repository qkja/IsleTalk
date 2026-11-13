package http

import (
	"net/http"
)

// Healthz 存活探针（宪法 D 组）：进程活着即 200，MUST NOT 探测依赖可用性。
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ok(map[string]string{"status": "ok"}))
}

// Readyz 就绪探针（宪法 D 组）：可接流量即 200；依赖可用性由调用方注入的 checker 判定。
// 当前骨架未接依赖探测，保持最小实现，待基础设施装配时注入 MongoDB/Redis ping。
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ok(map[string]string{"status": "ready"}))
}

// Startupz 启动探针（宪法 D 组）：启动完成即 200，仅在启动阶段使用。
// 与存活/就绪语义分离（存活=进程活着，就绪=可接流量，启动=启动完成）。
func (h *Handler) Startupz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ok(map[string]string{"status": "started"}))
}
