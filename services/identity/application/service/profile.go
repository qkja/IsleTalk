package service

import (
	"context"
	"log/slog"

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
	slog.Info("update profile", "user_id", userID.String())
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("update profile find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.UpdateProfile(nickname, avatar, signature)
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("update profile save failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	return u, nil
}

// SetVisibility 设置资料可见范围（FR-018）。
func (s *ProfileService) SetVisibility(ctx context.Context, userID valueobject.UserId, v entity.Visibility) (entity.User, error) {
	slog.Info("set visibility", "user_id", userID.String(), "visibility", string(v))
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("set visibility find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.SetVisibility(v)
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("set visibility save failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	return u, nil
}

// SetAllowStrangerAdd 设置是否允许陌生人添加（FR-018）。
func (s *ProfileService) SetAllowStrangerAdd(ctx context.Context, userID valueobject.UserId, allow bool) (entity.User, error) {
	slog.Info("set allow stranger add", "user_id", userID.String(), "allow", allow)
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("set allow stranger add find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.SetAllowStrangerAdd(allow)
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("set allow stranger add save failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	return u, nil
}

// SetLocale 设置界面语言（FR-019，zh-CN / en-US）。
func (s *ProfileService) SetLocale(ctx context.Context, userID valueobject.UserId, locale string) (entity.User, error) {
	slog.Info("set locale", "user_id", userID.String(), "locale", locale)
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("set locale find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.SetLocale(locale)
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("set locale save failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	return u, nil
}
