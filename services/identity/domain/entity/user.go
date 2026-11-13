package entity

import (
	"errors"
	"time"

	"isletalk/identity/domain/valueobject"
)

// UserStatus 是账号生命周期状态（data-model：active → deactivated）。
type UserStatus string

const (
	UserActive      UserStatus = "active"
	UserDeactivated UserStatus = "deactivated"
)

// Visibility 是资料可见范围（FR-018）。
type Visibility string

const (
	VisibilityPublic      Visibility = "public"
	VisibilityFriendsOnly Visibility = "friends_only"
	VisibilityPrivate     Visibility = "private"
)

// User 是用户账号聚合根。
type User struct {
	id               valueobject.UserId
	publicName       valueobject.PublicName
	phoneHash        valueobject.PhoneHash
	passwordHash     valueobject.PasswordHash
	nickname         string
	avatarURL        string
	signature        string
	visibility       Visibility
	allowStrangerAdd bool
	locale           string
	status           UserStatus
	createdAt        time.Time
}

var ErrInvalidStatus = errors.New("invalid user status")

// NewUser 创建新用户账号。手机号哈希唯一、公开账号名唯一由仓储层保证。
func NewUser(
	id valueobject.UserId,
	publicName valueobject.PublicName,
	phoneHash valueobject.PhoneHash,
	locale string,
) (User, error) {
	if locale == "" {
		locale = "zh-CN"
	}
	return User{
		id:               id,
		publicName:       publicName,
		phoneHash:        phoneHash,
		visibility:       VisibilityFriendsOnly,
		allowStrangerAdd: true,
		locale:           locale,
		status:           UserActive,
		createdAt:        time.Now().UTC(),
	}, nil
}

// ID 返回不可变内部账号 ID。
func (u User) ID() valueobject.UserId { return u.id }

// PublicName 返回公开账号名。
func (u User) PublicName() valueobject.PublicName { return u.publicName }

// PhoneHash 返回手机号哈希。
func (u User) PhoneHash() valueobject.PhoneHash { return u.phoneHash }

// PasswordHash 返回密码哈希（登录校验用）。
func (u User) PasswordHash() valueobject.PasswordHash { return u.passwordHash }

// SetPassword 设置密码哈希（注册或重置时）。
func (u *User) SetPassword(h valueobject.PasswordHash) { u.passwordHash = h }

// Nickname 返回昵称。
func (u User) Nickname() string { return u.nickname }

// AvatarURL 返回头像。
func (u User) AvatarURL() string { return u.avatarURL }

// Signature 返回个性签名。
func (u User) Signature() string { return u.signature }

// Visibility 返回资料可见范围。
func (u User) Visibility() Visibility { return u.visibility }

// AllowStrangerAdd 返回是否允许陌生人添加。
func (u User) AllowStrangerAdd() bool { return u.allowStrangerAdd }

// Locale 返回界面语言。
func (u User) Locale() string { return u.locale }

// Status 返回账号状态。
func (u User) Status() UserStatus { return u.status }

// CreatedAt 返回注册时间。
func (u User) CreatedAt() time.Time { return u.createdAt }

// UpdateProfile 更新昵称/头像/签名（FR-017）。返回不可变语义的新状态由调用方决定。
func (u *User) UpdateProfile(nickname, avatarURL, signature string) {
	u.nickname = nickname
	u.avatarURL = avatarURL
	u.signature = signature
}

// SetVisibility 设置资料可见范围（FR-018）。
func (u *User) SetVisibility(v Visibility) {
	switch v {
	case VisibilityPublic, VisibilityFriendsOnly, VisibilityPrivate:
		u.visibility = v
	default:
		u.visibility = VisibilityFriendsOnly
	}
}

// SetAllowStrangerAdd 设置是否允许陌生人添加（FR-018）。
func (u *User) SetAllowStrangerAdd(allow bool) { u.allowStrangerAdd = allow }

// SetPublicName 修改公开账号名（FR-014）。内部 ID 不变。
func (u *User) SetPublicName(n valueobject.PublicName) { u.publicName = n }

// SetLocale 设置界面语言（FR-019）。
func (u *User) SetLocale(locale string) { u.locale = locale }

// Deactivate 注销账号（FR-027/028）。
func (u *User) Deactivate() { u.status = UserDeactivated }

// IsActive 判断账号是否活跃。
func (u User) IsActive() bool { return u.status == UserActive }
