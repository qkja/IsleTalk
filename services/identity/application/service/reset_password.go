package service

import (
	"context"
	"errors"
	"log/slog"

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
	slog.Info("reset password", "phone_hash", phoneHash.String())

	ok, err := s.codes.Verify(ctx, phoneHash, code)
	if err != nil {
		slog.Error("reset password verify code failed", "err", err, "phone_hash", phoneHash.String())
		return err
	}
	if !ok {
		slog.Warn("reset password invalid code", "phone_hash", phoneHash.String())
		return ErrResetFailed
	}

	user, err := s.users.FindByPhoneHash(ctx, phoneHash)
	if err != nil {
		if port.IsNotFound(err) {
			slog.Warn("reset password user not found", "phone_hash", phoneHash.String())
			return ErrResetFailed
		}
		slog.Error("reset password find user failed", "err", err, "phone_hash", phoneHash.String())
		return err
	}

	hash, err := s.passwords.Hash(newPassword)
	if err != nil {
		slog.Error("reset password hash failed", "err", err)
		return err
	}
	user.SetPassword(hash)
	if err := s.users.Update(ctx, user); err != nil {
		slog.Error("reset password update failed", "err", err, "user_id", user.ID().String())
		return err
	}

	// 旧会话全失效（FR-004）
	if err := s.sessions.RevokeAllByUser(ctx, user.ID()); err != nil {
		slog.Error("reset password revoke sessions failed", "err", err, "user_id", user.ID().String())
		return err
	}
	slog.Info("reset password success", "user_id", user.ID().String())
	return nil
}
