package http

import (
	"encoding/json"
	"net/http"

	"isletalk/relationship/application/service"
	"isletalk/relationship/domain/valueobject"
)

// Handler 是客户端信封接口（T034）：发申请/接受/备注/拉黑/好友列表/删除/申请列表。
type Handler struct {
	friendRequest *service.FriendRequestService
	respond       *service.RespondService
	remark        *service.RemarkService
	block         *service.BlockService
	friendList    *service.FriendListService
}

func NewHandler(
	friendRequest *service.FriendRequestService,
	respond *service.RespondService,
	remark *service.RemarkService,
	block *service.BlockService,
	friendList *service.FriendListService,
) *Handler {
	return &Handler{friendRequest: friendRequest, respond: respond, remark: remark, block: block, friendList: friendList}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/friend-request", h.SendRequest)
	mux.HandleFunc("POST /api/v1/friend-respond", h.Respond)
	mux.HandleFunc("POST /api/v1/remark", h.SetRemark)
	mux.HandleFunc("POST /api/v1/block", h.SetBlock)
	mux.HandleFunc("GET /api/v1/friends", h.ListFriends)
	mux.HandleFunc("DELETE /api/v1/friend", h.DeleteFriend)
	mux.HandleFunc("GET /healthz", h.Healthz)
}

// SendRequest 发好友申请（FR-002）。
func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InitiatorID string `json:"initiatorId"`
		TargetID    string `json:"targetId"`
		Message     string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	initiator, err := valueobject.NewUserId(req.InitiatorID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "发起方标识无效"))
		return
	}
	target, err := valueobject.NewUserId(req.TargetID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "目标标识无效"))
		return
	}
	got, err := h.friendRequest.Send(r.Context(), initiator, target, req.Message)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(map[string]interface{}{"status": got.Status()}))
}

// Respond 接受/拒绝好友申请（FR-002/004）。
func (h *Handler) Respond(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InitiatorID string `json:"initiatorId"`
		TargetID    string `json:"targetId"`
		Accept      bool   `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	initiator, _ := valueobject.NewUserId(req.InitiatorID)
	target, _ := valueobject.NewUserId(req.TargetID)
	if _, err := h.respond.Respond(r.Context(), initiator, target, req.Accept); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// SetRemark 设置好友备注（FR-005）。
func (h *Handler) SetRemark(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OwnerID  string `json:"ownerId"`
		FriendID string `json:"friendId"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	owner, _ := valueobject.NewUserId(req.OwnerID)
	friend, _ := valueobject.NewUserId(req.FriendID)
	if _, err := h.remark.SetRemark(r.Context(), owner, friend, req.Content); err != nil {
		writeJSON(w, http.StatusBadRequest, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// SetBlock 拉黑/解除拉黑（FR-008）。
func (h *Handler) SetBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OwnerID  string `json:"ownerId"`
		TargetID string `json:"targetId"`
		Blocked  bool   `json:"blocked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	owner, _ := valueobject.NewUserId(req.OwnerID)
	target, _ := valueobject.NewUserId(req.TargetID)
	if err := h.block.SetBlock(r.Context(), owner, target, req.Blocked); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(nil))
}

// ListFriends 好友列表（FR-006，分页）。
func (h *Handler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	uid, err := valueobject.NewUserId(userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "用户标识无效"))
		return
	}
	friends, err := h.friendList.List(r.Context(), uid, 0, 50)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err))
		return
	}
	writeJSON(w, http.StatusOK, ok(map[string]interface{}{"friends": friends}))
}

// DeleteFriend 删除好友（FR-007，双向解除）。
func (h *Handler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"userId"`
		FriendID string `json:"friendId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fail("ARG_INVALID", "参数错误"))
		return
	}
	uid, _ := valueobject.NewUserId(req.UserID)
	fid, _ := valueobject.NewUserId(req.FriendID)
	if err := h.friendList.Delete(r.Context(), uid, fid); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorEnvelope(err))
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
