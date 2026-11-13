package service

import (
	"context"
	"errors"
	"log/slog"

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
	slog.Info("list sessions", "user_id", userID.String())
	res, err := s.sessions.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("list sessions failed", "err", err, "user_id", userID.String())
	}
	return res, err
}

// Revoke 撤销指定会话（远程下线）。
func (s *SessionService) Revoke(ctx context.Context, sessionID string) error {
	slog.Info("revoke session", "session_id", sessionID)
	if err := s.sessions.Revoke(ctx, sessionID); err != nil {
		slog.Error("revoke session failed", "err", err, "session_id", sessionID)
		return err
	}
	return nil
}

// RevokeAllByUser 撤销用户全部会话（改密/注销全失效）。
func (s *SessionService) RevokeAllByUser(ctx context.Context, userID valueobject.UserId) error {
	slog.Info("revoke all sessions", "user_id", userID.String())
	if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
		slog.Error("revoke all sessions failed", "err", err, "user_id", userID.String())
		return err
	}
	return nil
}

// Validate 验证会话是否有效并续期（FR-011）：撤销后立即拒绝，活跃时更新最近活跃时间。
func (s *SessionService) Validate(ctx context.Context, sessionID string) (entity.Session, error) {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		if port.IsNotFound(err) {
			slog.Warn("session validate not found", "session_id", sessionID)
			return entity.Session{}, ErrSessionRevoked
		}
		slog.Error("session validate find failed", "err", err, "session_id", sessionID)
		return entity.Session{}, err
	}
	if sess.Revoked() {
		slog.Warn("session validate revoked", "session_id", sessionID)
		return entity.Session{}, ErrSessionRevoked
	}
	// 续期：更新最近活跃时间（FR-011）
	if err := s.sessions.Touch(ctx, sessionID); err != nil {
		slog.Warn("session touch failed", "err", err, "session_id", sessionID)
	}
	return sess, nil
}
