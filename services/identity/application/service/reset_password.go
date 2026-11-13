package service

import (
	"context"
	"errors"

	"isletalk/identity/domain/port"
)

// ResetPasswordService 是找回密码用例（FR-004）。
type ResetPasswordService struct {
	users     port.UserRepository
	sessions  port.SessionRepository
	codes     VerificationCodeStore
	phones    PhoneHasher
	passwords PasswordHasher
}

func NewResetPasswordService(
	users port.UserRepository,
	sessions port.SessionRepository,
	codes VerificationCodeStore,
	phones PhoneHasher,
	passwords PasswordHasher,
) *ResetPasswordService {
	return &ResetPasswordService{
		users:     users,
		sessions:  sessions,
		codes:     codes,
		phones:    phones,
		passwords: passwords,
	}
}

var ErrResetFailed = errors.New("reset password failed")

// Reset 通过手机号验证后重置密码，旧会话全失效（FR-004）。
func (s *ResetPasswordService) Reset(ctx context.Context, phone, code, newPassword string) error {
	phoneHash := s.phones.Hash(phone)

	ok, err := s.codes.Verify(ctx, phoneHash, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrResetFailed
	}

	user, err := s.users.FindByPhoneHash(ctx, phoneHash)
	if err != nil {
		if port.IsNotFound(err) {
			return ErrResetFailed
		}
		return err
	}

	hash, err := s.passwords.Hash(newPassword)
	if err != nil {
		return err
	}
	user.SetPassword(hash)
	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	// 旧会话全失效（FR-004）
	return s.sessions.RevokeAllByUser(ctx, user.ID())
}
