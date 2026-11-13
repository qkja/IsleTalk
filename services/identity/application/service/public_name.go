package service

import (
	"context"
	"errors"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/service"
	"isletalk/identity/domain/valueobject"
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
	n, err := valueobject.NewPublicName(name)
	if err != nil {
		return entity.User{}, err
	}
	available, err := s.uniqueness.IsAvailable(ctx, n)
	if err != nil {
		return entity.User{}, err
	}
	if !available {
		return entity.User{}, ErrNameUnavailable
	}
	if err := s.claims.Claim(ctx, n, userID); err != nil {
		return entity.User{}, err
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}
	u.SetPublicName(n)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}

// Change 修改公开账号名（FR-014）：旧名进入保留期，内部 ID 不变。
func (s *PublicNameService) Change(ctx context.Context, userID valueobject.UserId, newName string) (entity.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return entity.User{}, err
	}

	n, err := valueobject.NewPublicName(newName)
	if err != nil {
		return entity.User{}, err
	}
	available, err := s.uniqueness.IsAvailable(ctx, n)
	if err != nil {
		return entity.User{}, err
	}
	if !available {
		return entity.User{}, ErrNameUnavailable
	}

	// 释放旧名（进入保留期），占用新名
	if u.PublicName().String() != "" {
		if err := s.claims.Release(ctx, u.PublicName()); err != nil {
			return entity.User{}, err
		}
	}
	if err := s.claims.Claim(ctx, n, userID); err != nil {
		return entity.User{}, err
	}

	u.SetPublicName(n)
	if err := s.users.Update(ctx, u); err != nil {
		return entity.User{}, err
	}
	return u, nil
}
