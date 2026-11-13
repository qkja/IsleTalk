package service

import (
	"context"

	"isletalk/identity/domain/valueobject"
)

// PasswordHasher 哈希/校验密码（Argon2id，由 infrastructure 实现）。
type PasswordHasher interface {
	Hash(plain string) (valueobject.PasswordHash, error)
	Verify(plain string, hash valueobject.PasswordHash) bool
}

// PhoneHasher 将手机号哈希为 PhoneHash（加盐哈希，由 infrastructure 实现）。
type PhoneHasher interface {
	Hash(phone string) valueobject.PhoneHash
}

// VerificationCodeStore 存储/校验/删除验证码（一次性，由 infrastructure 用 Redis 实现）。
type VerificationCodeStore interface {
	// Save 保存验证码（带 TTL）。
	Save(ctx context.Context, phone valueobject.PhoneHash, code string) error
	// Verify 校验验证码，成功后删除（一次性语义）。
	Verify(ctx context.Context, phone valueobject.PhoneHash, code string) (bool, error)
}

// IDGenerator 生成不可变内部账号 ID（UUIDv7，由 infrastructure 实现）。
type IDGenerator interface {
	NewID() valueobject.UserId
}
