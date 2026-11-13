package service

import (
	"context"
	"errors"
	"log/slog"

	"isletalk/identity/domain/entity"
	"isletalk/identity/domain/port"
	"isletalk/identity/domain/valueobject"
)

// 用例错误（对外统一错误码，避免泄露注册状态）。
var (
	ErrInvalidCode       = errors.New("invalid or expired verification code")
	ErrPhoneTaken        = errors.New("phone already registered") // 仅内部使用；对外不区分
	ErrInvalidCredential = errors.New("invalid credential")
)

// RegisterService 是注册用例（FR-001/002）。
// 防枚举：对外不区分"手机号已注册"与"验证码错误"（SC-003）。
type RegisterService struct {
	users     port.UserRepository
	sessions  port.SessionRepository
	tokens    port.TokenIssuer
	codes     VerificationCodeStore
	phones    PhoneHasher
	passwords PasswordHasher
	ids       IDGenerator
}

func NewRegisterService(
	users port.UserRepository,
	sessions port.SessionRepository,
	tokens port.TokenIssuer,
	codes VerificationCodeStore,
	phones PhoneHasher,
	passwords PasswordHasher,
	ids IDGenerator,
) *RegisterService {
	return &RegisterService{
		users:     users,
		sessions:  sessions,
		tokens:    tokens,
		codes:     codes,
		phones:    phones,
		passwords: passwords,
		ids:       ids,
	}
}

// RegisterInput 是注册命令。
type RegisterInput struct {
	Phone      string // 明文手机号（仅用于哈希，随即弃）
	Code       string // 验证码
	Password   string // 明文密码（仅用于哈希，随即弃）
	PublicName string // 公开账号名
	DeviceID   string
}

// RegisterResult 是注册结果。
type RegisterResult struct {
	User      entity.User
	SessionID string
	Tokens    port.TokenPair
}

// Register 执行注册：校验验证码 → 建账号 → 建会话 → 发令牌。
func (s *RegisterService) Register(ctx context.Context, in RegisterInput) (RegisterResult, error) {
	phoneHash := s.phones.Hash(in.Phone)
	slog.Info("register", "phone_hash", phoneHash.String())

	// 校验验证码（一次性）
	ok, err := s.codes.Verify(ctx, phoneHash, in.Code)
	if err != nil {
		slog.Error("register verify code failed", "err", err, "phone_hash", phoneHash.String())
		return RegisterResult{}, err
	}
	if !ok {
		// 防枚举：验证码错误与手机号已注册统一为 ErrInvalidCode，不区分
		slog.Warn("register invalid code", "phone_hash", phoneHash.String())
		return RegisterResult{}, ErrInvalidCode
	}

	// 手机号已注册则拒绝（内部知道原因，对外统一）
	if _, err := s.users.FindByPhoneHash(ctx, phoneHash); err == nil {
		slog.Info("register phone already taken", "phone_hash", phoneHash.String())
		return RegisterResult{}, ErrPhoneTaken
	} else if !port.IsNotFound(err) {
		slog.Error("register find by phone failed", "err", err, "phone_hash", phoneHash.String())
		return RegisterResult{}, err
	}

	publicName, err := valueobject.NewPublicName(in.PublicName)
	if err != nil {
		slog.Error("register invalid public name", "err", err)
		return RegisterResult{}, err
	}
	passwordHash, err := s.passwords.Hash(in.Password)
	if err != nil {
		slog.Error("register hash password failed", "err", err)
		return RegisterResult{}, err
	}

	user, err := entity.NewUser(s.ids.NewID(), publicName, phoneHash, "zh-CN")
	if err != nil {
		slog.Error("register new user failed", "err", err)
		return RegisterResult{}, err
	}
	user.SetPassword(passwordHash)

	created, err := s.users.Create(ctx, user)
	if err != nil {
		slog.Error("register create user failed", "err", err, "user_id", created.ID().String())
		return RegisterResult{}, err
	}

	session := entity.NewSession(s.ids.NewID().String(), created.ID(), in.DeviceID)
	savedSession, err := s.sessions.Create(ctx, session)
	if err != nil {
		slog.Error("register create session failed", "err", err, "user_id", created.ID().String())
		return RegisterResult{}, err
	}

	tokens, err := s.tokens.Issue(ctx, created.ID())
	if err != nil {
		slog.Error("register issue token failed", "err", err, "user_id", created.ID().String())
		return RegisterResult{}, err
	}

	slog.Info("register success", "user_id", created.ID().String())
	return RegisterResult{User: created, SessionID: savedSession.ID(), Tokens: tokens}, nil
}
