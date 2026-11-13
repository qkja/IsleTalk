package service

import (
	"context"
	"log/slog"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// DeactivateService 是注销用例（FR-027/028）。
type DeactivateService struct {
	users         port.UserRepository
	sessions      port.SessionRepository
	verifications port.VerificationRepository
}

func NewDeactivateService(
	users port.UserRepository,
	sessions port.SessionRepository,
	verifications port.VerificationRepository,
) *DeactivateService {
	return &DeactivateService{users: users, sessions: sessions, verifications: verifications}
}

// Deactivate 注销账号：状态置 deactivated，全部会话失效，撤回实名（FR-027/028）。
// 二次确认与重新验证身份由上层（api 层）在执行前完成（FR-027）。
func (s *DeactivateService) Deactivate(ctx context.Context, userID valueobject.UserId) (entity.User, error) {
	slog.Info("deactivate account", "user_id", userID.String())
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("deactivate find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.Deactivate()
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("deactivate update user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	// 全部会话失效（FR-028）
	if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
		slog.Error("deactivate revoke sessions failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	// 撤回实名、删除身份信息（FR-025/028）
	if err := s.verifications.Revoke(ctx, userID); err != nil {
		slog.Error("deactivate revoke verification failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	slog.Info("deactivate account success", "user_id", userID.String())
	return u, nil
}
