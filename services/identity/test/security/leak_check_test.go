package security

import (
	"strings"
	"testing"

	"isletalk/identity/infrastructure/auth"
)

// 零泄露验证（SC-007）：身份信息/手机号明文 MUST NOT 出现在哈希值、令牌、
// 或任何字符串化输出中。此处验证核心适配器的不可逆性。
func TestPhoneHashDoesNotLeakPlaintext(t *testing.T) {
	plain := "13800001111"
	h := auth.NewPhoneHasher().Hash(plain)
	if strings.Contains(h.String(), plain) {
		t.Fatalf("phone hash leaks plaintext: %q contains %q", h.String(), plain)
	}
}

func TestPasswordHashDoesNotLeakPlaintext(t *testing.T) {
	plain := "correct horse battery staple"
	h, err := auth.NewPasswordHasher().Hash(plain)
	if err != nil {
		t.Fatalf("Hash err: %v", err)
	}
	if strings.Contains(h.String(), plain) {
		t.Fatalf("password hash leaks plaintext")
	}
}

func TestIDIsUnpredictable(t *testing.T) {
	g := auth.NewIDGenerator()
	a := g.NewID().String()
	b := g.NewID().String()
	if a == b {
		t.Fatal("two generated IDs must differ")
	}
	// 内部账号 ID MUST NOT 为自增序号（data-model）
	if len(a) < 8 {
		t.Fatalf("ID too short: %q", a)
	}
}
