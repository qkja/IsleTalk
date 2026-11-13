package valueobject

import "strings"

// reservedWords 是公开账号名的保留词（FR-013）。
// MUST NOT 被普通用户占用；清单在 plan/data-model 锁定并对用户可见。
var reservedWords = map[string]struct{}{
	"admin":         {},
	"administrator": {},
	"root":          {},
	"system":        {},
	"moderator":     {},
	"official":      {},
	"support":       {},
	"help":          {},
	"service":       {},
	"staff":         {},
	"bot":           {},
	"isletalk":      {},
	"官方":            {},
	"管理员":           {},
	"系统":            {},
	"客服":            {},
	"机器人":           {},
}

// isReserved 判断账号名（英文大小写不敏感）是否为保留词。
func isReserved(name string) bool {
	_, ok := reservedWords[strings.ToLower(name)]
	return ok
}
