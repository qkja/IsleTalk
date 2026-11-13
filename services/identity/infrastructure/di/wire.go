package di

import (
	"isletalk/identity/application/service"
	"isletalk/identity/domain/port"
	"isletalk/identity/infrastructure/auth"
)

// Container 是组合根（宪法原则 V：组合根 MUST 唯一且位于最外层）。
// 依赖由构造函数显式注入，禁止服务定位器或全局单例。
type Container struct {
	Register     *service.RegisterService
	Login        *service.LoginService
	Reset        *service.ResetPasswordService
	Session      *service.SessionService
	PublicName   *service.PublicNameService
	Profile      *service.ProfileService
	Verification *service.VerificationService
	Deactivate   *service.DeactivateService
}

// NewContainer 装配全部 application 服务。
// repos 由调用方（main）注入具体实现（内存 mock / MongoDB）。
func NewContainer(
	users port.UserRepository,
	sessions port.SessionRepository,
	verifications port.VerificationRepository,
	nameClaims port.NameClaimRepository,
	tokens port.TokenIssuer,
	codes service.VerificationCodeStore,
) *Container {
	passwords := auth.NewPasswordHasher()
	phones := auth.NewPhoneHasher()
	ids := auth.NewIDGenerator()

	return &Container{
		Register: service.NewRegisterService(
			users, sessions, tokens, codes, phones, passwords, ids,
		),
		Login: service.NewLoginService(
			users, sessions, tokens, codes, phones, passwords, ids,
		),
		Reset: service.NewResetPasswordService(
			users, sessions, codes, phones, passwords,
		),
		Session:      service.NewSessionService(sessions),
		PublicName:   service.NewPublicNameService(users, nameClaims),
		Profile:      service.NewProfileService(users),
		Verification: service.NewVerificationService(verifications),
		Deactivate:   service.NewDeactivateService(users, sessions, verifications),
	}
}
