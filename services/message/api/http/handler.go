package http

import (
	"encoding/json"
	"net/http"

	"isletalk/message/application/service"
	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

// Handler 是客户端信封接口（T024）：会话列表/置顶/免打扰/删除/标记未读。
type Handler struct {
	conversation *service.ConversationService
}

func NewHandler(conversation *service.ConversationService) *Handler {
	return &Handler{conversation: conversation}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/conversations", h.ListConversations)
	mux.HandleFunc("POST /api/v1/conversations/pin", h.Pin)
	mux.HandleFunc("POST /api/v1/conversations/mute", h.Mute)
	mux.HandleFunc("POST /api/v1/conversations/unread", h.MarkUnread)
	mux.HandleFunc("DELETE /api/v1/conversations", h.Delete)
	mux.HandleFunc("GET /healthz", h.Healthz)
}

// conversationDTO 是会话的对外 JSON 视图（领域模型与传输模型解耦，宪法原则 V）。
type conversationDTO struct {
	ConversationID     string   `json:"conversationId"`
	Type               string   `json:"type"`
	Participants       []string `json:"participants"`
	LastMessagePreview string   `json:"lastMessagePreview"`
	UnreadCount        int      `json:"unreadCount"`
	Pinned             bool     `json:"pinned"`
	Muted              bool     `json:"muted"`
	Deleted            bool     `json:"deleted"`
	LastActiveAt       int64    `json:"lastActiveAt"`
}

func toConversationDTO(c entity.Conversation) conversationDTO {
	parts := make([]string, 0, len(c.Participants()))
	for _, p := range c.Participants() {
		parts = append(parts, p.String())
	}
	return conversationDTO{
		ConversationID:     c.ConversationID(),
		Type:               string(c.Type()),
		Participants:       parts,
		LastMessagePreview: c.LastMessagePreview(),
		UnreadCount:        c.UnreadCount(),
		Pinned:             c.Pinned(),
		Muted:              c.Muted(),
		Deleted:            c.Deleted(),
		LastActiveAt:       c.LastActiveAt().UnixMilli(),
	}
}

// ListConversations 会话列表（FR-008，倒序/未读/摘要）。
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	uid, err := valueobject.NewUserId(r.URL.Query().Get("userId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	convs, err := h.conversation.List(r.Context(), uid, 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err))
		return
	}
	out := make([]conversationDTO, 0, len(convs))
	for _, c := range convs {
		out = append(out, toConversationDTO(c))
	}
	writeJSON(w, http.StatusOK, ok(map[string]interface{}{"conversations": out}))
}

// Pin 置顶/取消置顶（FR-024）。
func (h *Handler) Pin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConversationID string `json:"conversationId"`
		UserID         string `json:"userId"`
		Pinned         bool   `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	uid, err := valueobject.NewUserId(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	if err := h.conversation.Pin(r.Context(), req.ConversationID, uid, req.Pinned); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// Mute 免打扰/取消免打扰（FR-024）。
func (h *Handler) Mute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConversationID string `json:"conversationId"`
		UserID         string `json:"userId"`
		Muted          bool   `json:"muted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	uid, err := valueobject.NewUserId(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	if err := h.conversation.Mute(r.Context(), req.ConversationID, uid, req.Muted); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// MarkUnread 标记未读（FR-024）。
func (h *Handler) MarkUnread(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConversationID string `json:"conversationId"`
		UserID         string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	uid, err := valueobject.NewUserId(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	if err := h.conversation.MarkUnread(r.Context(), req.ConversationID, uid); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// Delete 删除会话（FR-024；对方发新消息时重新出现）。
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConversationID string `json:"conversationId"`
		UserID         string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	uid, err := valueobject.NewUserId(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	if err := h.conversation.Delete(r.Context(), req.ConversationID, uid); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// Healthz 存活探针（宪法 D 组）。
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ok(map[string]string{"status": "ok"}))
}

func writeJSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
