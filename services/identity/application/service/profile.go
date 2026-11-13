package service

import (
	"context"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// ProfileService 是资料与隐私用例（FR-017~019）。
type ProfileService struct {
	users port.UserRepository
}

func NewProfileService(users port.UserRepository) *ProfileService {
	return &ProfileService{users: users}
}

// UpdateProfile 更新昵称/头像/签名。
func (s *ProfileService) UpdateProfile(ctx context.Context, userID valueobject.UserId, nickname, avatar, signature string) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.UpdateProfile(nickname, avatar, signature)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}

// SetVisibility 设置资料可见范围（FR-018）。
func (s *ProfileService) SetVisibility(ctx context.Context, userID valueobject.UserId, v entity.Visibility) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.SetVisibility(v)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}

// SetAllowStrangerAdd 设置是否允许陌生人添加（FR-018）。
func (s *ProfileService) SetAllowStrangerAdd(ctx context.Context, userID valueobject.UserId, allow bool) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.SetAllowStrangerAdd(allow)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}

// SetLocale 设置界面语言（FR-019，zh-CN / en-US）。
func (s *ProfileService) SetLocale(ctx context.Context, userID valueobject.UserId, locale string) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.SetLocale(locale)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}
