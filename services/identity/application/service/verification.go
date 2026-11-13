package service

import (
	"context"
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
	v := entity.NewVerification(userID, idType, idDigest, consentAt, time.Now().Add(retention))
	return s.verifications.Save(ctx, v)
}

// Revoke 撤回实名（FR-024）：删除身份信息 + 信任降级 + 额度回落。
func (s *VerificationService) Revoke(ctx context.Context, userID valueobject.UserId) error {
	return s.verifications.Revoke(ctx, userID)
}
