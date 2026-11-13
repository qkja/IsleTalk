package auth

import (
	"crypto/rand"
	"fmt"

	"isletalk/identity/domain/valueobject"
)

// IDGenerator 生成不可变内部账号 ID（随机 UUID 形态，避免可枚举）。
// 依 data-model：MUST 用不可预测 UUIDv4/v7，MUST NOT 用自增序号。
type IDGenerator struct{}

func NewIDGenerator() IDGenerator { return IDGenerator{} }

func (IDGenerator) NewID() valueobject.UserId {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("crypto/rand unavailable: %v", err))
	}
	// UUIDv4 形态
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	id, _ := valueobject.NewUserId(fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16],
	))
	return id
}
