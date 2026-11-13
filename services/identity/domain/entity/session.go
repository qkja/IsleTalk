package entity

import (
	"time"

	"isletalk/identity/domain/valueobject"
)

// Session 是设备会话聚合根（FR-008~011）。
type Session struct {
	id           string
	userID       valueobject.UserId
	deviceID     string
	createdAt    time.Time
	lastActiveAt time.Time
	revoked      bool
}

// NewSession 创建新会话。
func NewSession(id string, userID valueobject.UserId, deviceID string) Session {
	now := time.Now().UTC()
	return Session{
		id:           id,
		userID:       userID,
		deviceID:     deviceID,
		createdAt:    now,
		lastActiveAt: now,
	}
}

// ID 返回会话标识。
func (s Session) ID() string { return s.id }

// UserID 返回所属用户。
func (s Session) UserID() valueobject.UserId { return s.userID }

// DeviceID 返回设备标识。
func (s Session) DeviceID() string { return s.deviceID }

// CreatedAt 返回登录时间。
func (s Session) CreatedAt() time.Time { return s.createdAt }

// LastActiveAt 返回最近活跃时间。
func (s Session) LastActiveAt() time.Time { return s.lastActiveAt }

// Revoked 返回是否已撤销。
func (s Session) Revoked() bool { return s.revoked }

// Touch 更新最近活跃时间。
func (s *Session) Touch() { s.lastActiveAt = time.Now().UTC() }

// Revoke 撤销会话。撤销后立即拒绝（FR-011）。
func (s *Session) Revoke() { s.revoked = true }

// IsActive 判断会话是否有效（未撤销）。
func (s Session) IsActive() bool { return !s.revoked }
