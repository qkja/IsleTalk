package service

import (
	"context"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
)

// LoginService 是登录用例（FR-003）。
// 防枚举：账号不存在与密码错误统一返回 ErrInvalidCredential（SC-003）。
type LoginService struct {
	users     port.UserRepository
	sessions  port.SessionRepository
	tokens    port.TokenIssuer
	phones    PhoneHasher
	passwords PasswordHasher
	ids       IDGenerator
}

func NewLoginService(
	users port.UserRepository,
	sessions port.SessionRepository,
	tokens port.TokenIssuer,
	phones PhoneHasher,
	passwords PasswordHasher,
	ids IDGenerator,
) *LoginService {
	return &LoginService{
		users:     users,
		sessions:  sessions,
		tokens:    tokens,
		phones:    phones,
		passwords: passwords,
		ids:       ids,
	}
}

// LoginInput 是登录命令。
type LoginInput struct {
	Phone    string
	Password string
	DeviceID string
}

// LoginResult 是登录结果。
type LoginResult struct {
	User      entity.User
	SessionID string
	Tokens    port.TokenPair
}

// Login 执行登录：查用户 → 校验密码 → 建会话 → 发令牌。
// 防枚举：未注册与密码错误统一返回 ErrInvalidCredential，不区分。
func (s *LoginService) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	phoneHash := s.phones.Hash(in.Phone)

	user, err := s.users.FindByPhoneHash(ctx, phoneHash)
	if err != nil {
		if port.IsNotFound(err) {
			return LoginResult{}, ErrInvalidCredential
		}
		return LoginResult{}, err
	}

	if !s.passwords.Verify(in.Password, user.PasswordHash()) {
		return LoginResult{}, ErrInvalidCredential
	}

	session := entity.NewSession(s.ids.NewID().String(), user.ID(), in.DeviceID)
	saved, err := s.sessions.Create(ctx, session)
	if err != nil {
		return LoginResult{}, err
	}

	tokens, err := s.tokens.Issue(ctx, user.ID())
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{User: user, SessionID: saved.ID(), Tokens: tokens}, nil
}
