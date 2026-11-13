package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/service"
	"isletalk/identity/domain/valueobject"
	"isletalk/identity/infrastructure/config"
)

var (
	ErrNameUnavailable = errors.New("public name unavailable")
	ErrNameCooldown    = errors.New("public name change in cooldown")
)

// PublicNameService 是公开账号名用例（FR-013~015）。
type PublicNameService struct {
	users      port.UserRepository
	claims     port.NameClaimRepository
	uniqueness *service.NameUniqueness
}

func NewPublicNameService(users port.UserRepository, claims port.NameClaimRepository) *PublicNameService {
	return &PublicNameService{
		users:      users,
		claims:     claims,
		uniqueness: service.NewNameUniqueness(claims, users),
	}
}

// Set 设置公开账号名（注册后首次设置）。
func (s *PublicNameService) Set(ctx context.Context, userID valueobject.UserId, name string) (entity.User, error) {
	slog.Info("set public name", "user_id", userID.String())
	n, err := valueobject.NewPublicName(name)
	if err != nil {
		slog.Error("set public name invalid name", "err", err)
		return entity.User{}, err
	}
	available, err := s.uniqueness.IsAvailable(ctx, n)
	if err != nil {
		slog.Error("set public name check availability failed", "err", err, "name", n.String())
		return entity.User{}, err
	}
	if !available {
		slog.Warn("set public name unavailable", "name", n.String())
		return entity.User{}, ErrNameUnavailable
	}
	if err := s.claims.Claim(ctx, n, userID); err != nil {
		slog.Error("set public name claim failed", "err", err, "name", n.String())
		return entity.User{}, err
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("set public name find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	u.SetPublicName(n)
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("set public name update failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	return u, nil
}

// Change 修改公开账号名（FR-014）：旧名进入保留期，内部 ID 不变，受改名冷却期限制。
func (s *PublicNameService) Change(ctx context.Context, userID valueobject.UserId, newName string) (entity.User, error) {
	slog.Info("change public name", "user_id", userID.String())
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		slog.Error("change public name find user failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}

	// 改名冷却期（FR-014）：上次改名后默认 30 天内不可再次改名
	cooldown := time.Duration(config.Days("IDENTITY_NAME_COOLDOWN_DAYS", 30)) * 24 * time.Hour
	if !u.LastPublicNameChangeAt().IsZero() && time.Since(u.LastPublicNameChangeAt()) < cooldown {
		slog.Warn("change public name in cooldown", "user_id", userID.String())
		return entity.User{}, ErrNameCooldown
	}

	n, err := valueobject.NewPublicName(newName)
	if err != nil {
		slog.Error("change public name invalid name", "err", err)
		return entity.User{}, err
	}
	available, err := s.uniqueness.IsAvailable(ctx, n)
	if err != nil {
		slog.Error("change public name check availability failed", "err", err, "name", n.String())
		return entity.User{}, err
	}
	if !available {
		slog.Warn("change public name unavailable", "name", n.String())
		return entity.User{}, ErrNameUnavailable
	}

	// 释放旧名（进入保留期），占用新名
	if u.PublicName().String() != "" {
		if err := s.claims.Release(ctx, u.PublicName()); err != nil {
			slog.Error("change public name release old failed", "err", err, "old_name", u.PublicName().String())
			return entity.User{}, err
		}
	}
	if err := s.claims.Claim(ctx, n, userID); err != nil {
		slog.Error("change public name claim failed", "err", err, "name", n.String())
		return entity.User{}, err
	}

	u.SetPublicName(n)
	u.MarkPublicNameChanged()
	if err := s.users.Update(ctx, u); err != nil {
		slog.Error("change public name update failed", "err", err, "user_id", userID.String())
		return entity.User{}, err
	}
	slog.Info("change public name success", "user_id", userID.String(), "name", n.String())
	return u, nil
}
