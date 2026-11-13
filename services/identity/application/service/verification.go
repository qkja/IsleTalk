package service

import (
	"context"
	"log/slog"
	"time"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// VerificationService 是实名登记用例（FR-020~025）。
// 首段切片只"收集 + 加密存储"，不比对权威数据源（FR-020）。
type VerificationService struct {
	verifications port.VerificationRepository
}

func NewVerificationService(v port.VerificationRepository) *VerificationService {
	return &VerificationService{verifications: v}
}

// Submit 提交实名登记（已在外部取得单独同意 FR-021）。
// idDigest 是 AEAD 加密后的身份摘要，MUST NOT 落明文（FR-022）。
func (s *VerificationService) Submit(
	ctx context.Context,
	userID valueobject.UserId,
	idType string,
	idDigest string,
	consentAt time.Time,
	retention time.Duration,
) (entity.Verification, error) {
	slog.Info("submit verification", "user_id", userID.String(), "id_type", idType)
	v := entity.NewVerification(userID, idType, idDigest, consentAt, time.Now().Add(retention))
	saved, err := s.verifications.Save(ctx, v)
	if err != nil {
		slog.Error("submit verification save failed", "err", err, "user_id", userID.String())
		return entity.Verification{}, err
	}
	return saved, nil
}

// Revoke 撤回实名（FR-024）：删除身份信息 + 信任降级 + 额度回落。
func (s *VerificationService) Revoke(ctx context.Context, userID valueobject.UserId) error {
	slog.Info("revoke verification", "user_id", userID.String())
	if err := s.verifications.Revoke(ctx, userID); err != nil {
		slog.Error("revoke verification failed", "err", err, "user_id", userID.String())
		return err
	}
	return nil
}

// PurgeExpired 删除保留期限到期的实名信息（FR-025 到期删除机制）。
func (s *VerificationService) PurgeExpired(ctx context.Context, now time.Time) (int64, error) {
	slog.Info("purge expired verifications", "before", now.UTC().String())
	n, err := s.verifications.DeleteExpired(ctx, now)
	if err != nil {
		slog.Error("purge expired verifications failed", "err", err)
		return 0, err
	}
	slog.Info("purge expired verifications done", "deleted", n)
	return n, nil
}
