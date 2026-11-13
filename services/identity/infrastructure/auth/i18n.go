package auth

// 非界面文案国际化（T047，FR-019）：验证码短信、错误提示、注销告知的 zh-CN/en-US 双语。
// 文案 MUST 外置为资源，MUST NOT 硬编码在业务逻辑中（宪法 G 组）。
var i18nMessages = map[string]map[string]string{
	"verify_code_sms": {
		"zh-CN": "您的验证码是 %s，%d 分钟内有效。",
		"en-US": "Your verification code is %s, valid for %d minutes.",
	},
	"error_invalid_code": {
		"zh-CN": "验证码错误或已过期",
		"en-US": "Invalid or expired verification code",
	},
	"error_invalid_credential": {
		"zh-CN": "账号或密码错误",
		"en-US": "Incorrect account or password",
	},
	"deactivate_notice": {
		"zh-CN": "账号注销后，数据将按保留策略处理。",
		"en-US": "After deactivation, your data will be handled per the retention policy.",
	},
}

// Message 返回指定 locale 的文案；缺失 locale 回退 zh-CN。
func Message(key, locale string) string {
	if m, ok := i18nMessages[key]; ok {
		if v, ok := m[locale]; ok {
			return v
		}
		return m["zh-CN"]
	}
	return key
}
