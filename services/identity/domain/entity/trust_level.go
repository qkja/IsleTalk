package entity

import "isletalk/identity/domain/valueobject"

// TrustLevel 是由实名登记派生的用户属性（FR-026）。
// 首段切片按"已提交"授予，不基于真实验证。
type TrustLevel struct {
	userID        valueobject.UserId
	level         int
	quotaOverride map[string]int // 受限维度上的放宽额度
}

// NewTrustLevel 为已提交实名的用户创建信任等级。
func NewTrustLevel(userID valueobject.UserId, level int, quota map[string]int) TrustLevel {
	if quota == nil {
		quota = map[string]int{}
	}
	return TrustLevel{userID: userID, level: level, quotaOverride: quota}
}

// UserID 返回所属用户。
func (t TrustLevel) UserID() valueobject.UserId { return t.userID }

// Level 返回信任等级。
func (t TrustLevel) Level() int { return t.level }

// Quota 返回放宽额度。
func (t TrustLevel) Quota() map[string]int { return t.quotaOverride }

// Downgrade 撤回实名后降级（FR-024）：等级归零、额度回落。
func (t *TrustLevel) Downgrade() {
	t.level = 0
	t.quotaOverride = map[string]int{}
}
