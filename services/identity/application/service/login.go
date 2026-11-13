package service

import (
	"context"
	"log/slog"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
)

// LoginService 是登录用例（FR-003）。
// 防枚举：账号不存在与密码错误统一返回 ErrInvalidCredential（SC-003）。
type LoginService struct {
	users     port.UserRepository
	sessions  port.SessionRepository
	tokens    port.TokenIssuer
	codes     VerificationCodeStore
	phones    PhoneHasher
	passwords PasswordHasher
	ids       IDGenerator
}

func NewLoginService(
	users port.UserRepository,
	sessions port.SessionRepository,
	tokens port.TokenIssuer,
	codes VerificationCodeStore,
	phones PhoneHasher,
	passwords PasswordHasher,
	ids IDGenerator,
) *LoginService {
	return &LoginService{
		users:     users,
		sessions:  sessions,
		tokens:    tokens,
		codes:     codes,
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

// LoginByCodeInput 是验证码登录命令（FR-003 辅登录）。
type LoginByCodeInput struct {
	Phone    string
	Code     string
	DeviceID string
}

// Login 执行登录：查用户 → 校验密码 → 建会话 → 发令牌。
// 防枚举：未注册与密码错误统一返回 ErrInvalidCredential，不区分。
func (s *LoginService) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	phoneHash := s.phones.Hash(in.Phone)
	slog.Info("login", "phone_hash", phoneHash.String())

	user, err := s.users.FindByPhoneHash(ctx, phoneHash)
	if err != nil {
		if port.IsNotFound(err) {
			// 防枚举：未注册与密码错误统一返回，不区分
			slog.Warn("login invalid credential", "phone_hash", phoneHash.String())
			return LoginResult{}, ErrInvalidCredential
		}
		slog.Error("login find by phone failed", "err", err, "phone_hash", phoneHash.String())
		return LoginResult{}, err
	}

	if !s.passwords.Verify(in.Password, user.PasswordHash()) {
		slog.Warn("login invalid credential", "user_id", user.ID().String())
		return LoginResult{}, ErrInvalidCredential
	}

	session := entity.NewSession(s.ids.NewID().String(), user.ID(), in.DeviceID)
	saved, err := s.sessions.Create(ctx, session)
	if err != nil {
		slog.Error("login create session failed", "err", err, "user_id", user.ID().String())
		return LoginResult{}, err
	}

	tokens, err := s.tokens.Issue(ctx, user.ID())
	if err != nil {
		slog.Error("login issue token failed", "err", err, "user_id", user.ID().String())
		return LoginResult{}, err
	}

	slog.Info("login success", "user_id", user.ID().String())
	return LoginResult{User: user, SessionID: saved.ID(), Tokens: tokens}, nil
}

// LoginByCode 通过手机号验证码登录（FR-003 辅登录）。
// 防枚举：验证码错误与账号不存在统一返回 ErrInvalidCredential，与密码登录一致（SC-003）。
func (s *LoginService) LoginByCode(ctx context.Context, in LoginByCodeInput) (LoginResult, error) {
	phoneHash := s.phones.Hash(in.Phone)
	slog.Info("login by code", "phone_hash", phoneHash.String())

	ok, err := s.codes.Verify(ctx, phoneHash, in.Code)
	if err != nil {
		slog.Error("login by code verify failed", "err", err, "phone_hash", phoneHash.String())
		return LoginResult{}, err
	}
	if !ok {
		slog.Warn("login by code invalid code", "phone_hash", phoneHash.String())
		return LoginResult{}, ErrInvalidCredential
	}

	user, err := s.users.FindByPhoneHash(ctx, phoneHash)
	if err != nil {
		if port.IsNotFound(err) {
			slog.Warn("login by code invalid credential", "phone_hash", phoneHash.String())
			return LoginResult{}, ErrInvalidCredential
		}
		slog.Error("login by code find by phone failed", "err", err, "phone_hash", phoneHash.String())
		return LoginResult{}, err
	}

	session := entity.NewSession(s.ids.NewID().String(), user.ID(), in.DeviceID)
	saved, err := s.sessions.Create(ctx, session)
	if err != nil {
		slog.Error("login by code create session failed", "err", err, "user_id", user.ID().String())
		return LoginResult{}, err
	}

	tokens, err := s.tokens.Issue(ctx, user.ID())
	if err != nil {
		slog.Error("login by code issue token failed", "err", err, "user_id", user.ID().String())
		return LoginResult{}, err
	}

	slog.Info("login by code success", "user_id", user.ID().String())
	return LoginResult{User: user, SessionID: saved.ID(), Tokens: tokens}, nil
}
