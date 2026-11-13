package service

import (
	"context"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// DeactivateService 是注销用例（FR-027/028）。
type DeactivateService struct {
	users    port.UserRepository
	sessions port.SessionRepository
}

func NewDeactivateService(users port.UserRepository, sessions port.SessionRepository) *DeactivateService {
	return &DeactivateService{users: users, sessions: sessions}
}

// Deactivate 注销账号：状态置 deactivated，全部会话失效。
// 二次确认与重新验证身份由上层（api 层）在执行前完成（FR-027）。
func (s *DeactivateService) Deactivate(ctx context.Context, userID valueobject.UserId) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.Deactivate()
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	// 全部会话失效（FR-028）
	if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
		return entity.User{}, err
	}
	return u, nil
}
