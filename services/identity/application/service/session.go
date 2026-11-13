package service

import (
	"context"
	"errors"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

var ErrSessionRevoked = errors.New("session revoked")

// SessionService 是会话用例（FR-008~011）。
type SessionService struct {
	sessions port.SessionRepository
}

func NewSessionService(sessions port.SessionRepository) *SessionService {
	return &SessionService{sessions: sessions}
}

// List 列出用户全部会话。
func (s *SessionService) List(ctx context.Context, userID valueobject.UserId) ([]entity.Session, error) {
	return s.sessions.ListByUser(ctx, userID)
}

// Revoke 撤销指定会话（远程下线）。
func (s *SessionService) Revoke(ctx context.Context, sessionID string) error {
	return s.sessions.Revoke(ctx, sessionID)
}

// RevokeAllByUser 撤销用户全部会话（改密/注销全失效）。
func (s *SessionService) RevokeAllByUser(ctx context.Context, userID valueobject.UserId) error {
	return s.sessions.RevokeAllByUser(ctx, userID)
}
