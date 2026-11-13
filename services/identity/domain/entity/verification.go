package entity

import (
	"time"

	"isletalk/identity/domain/valueobject"
)

// VerificationStatus 是实名登记状态（clarify 后：登记而非认证，submitted 而非 verified）。
type VerificationStatus string

const (
	VerificationPending   VerificationStatus = "pending"
	VerificationSubmitted VerificationStatus = "submitted"
	VerificationRevoked   VerificationStatus = "revoked"
)

// Verification 是实名登记聚合根（FR-020~025）。
// 首段切片只"收集 + 加密存储"，不比对权威数据源（FR-020）。
type Verification struct {
	userID         valueobject.UserId
	idType         string
	idDigest       string // AEAD 加密后的身份摘要，MUST NOT 落明文
	status         VerificationStatus
	consentAt      time.Time // 单独同意时间（FR-021）
	retentionUntil time.Time // 保留期限（FR-025）
}

// NewVerification 创建实名登记（已取得单独同意）。
func NewVerification(
	userID valueobject.UserId,
	idType string,
	idDigest string,
	consentAt time.Time,
	retentionUntil time.Time,
) Verification {
	return Verification{
		userID:         userID,
		idType:         idType,
		idDigest:       idDigest,
		status:         VerificationSubmitted,
		consentAt:      consentAt,
		retentionUntil: retentionUntil,
	}
}

// UserID 返回所属用户。
func (v Verification) UserID() valueobject.UserId { return v.userID }

// IDType 返回证件类型。
func (v Verification) IDType() string { return v.idType }

// IDDigest 返回加密后的身份摘要。
func (v Verification) IDDigest() string { return v.idDigest }

// Status 返回登记状态。
func (v Verification) Status() VerificationStatus { return v.status }

// ConsentAt 返回单独同意时间。
func (v Verification) ConsentAt() time.Time { return v.consentAt }

// RetentionUntil 返回保留期限。
func (v Verification) RetentionUntil() time.Time { return v.retentionUntil }

// Revoke 撤回实名（FR-024）：状态置 revoked，删除身份摘要。
func (v *Verification) Revoke() {
	v.status = VerificationRevoked
	v.idDigest = ""
}

// IsSubmitted 判断是否已提交且未撤回。
func (v Verification) IsSubmitted() bool { return v.status == VerificationSubmitted }
