package di

import (
	"isletalk/message/application/service"
	"isletalk/message/domain/port"
	"isletalk/message/infrastructure/cache"
)

// Container 是组合根（宪法原则 V：组合根 MUST 唯一且位于最外层）。
// 依赖由构造函数显式注入，禁止服务定位器或全局单例。
type Container struct {
	SendMessage  *service.SendMessageService
	ListMessages *service.ListMessagesService
	Recall       *service.RecallService
	MarkRead     *service.MarkReadService
	Conversation *service.ConversationService
}

// NewContainer 装配全部 application 服务。
// repos/client 由调用方（main）注入具体实现（内存 mock / MongoDB / gRPC）。
func NewContainer(
	messages port.MessageRepository,
	conversations port.ConversationRepository,
	relationship port.RelationshipClient,
) *Container {
	unread := cache.NewUnreadCache(nil)
	convCache := cache.NewConversationCache(nil)
	return &Container{
		SendMessage:  service.NewSendMessageService(messages, conversations, relationship, unread, convCache),
		ListMessages: service.NewListMessagesService(messages),
		Recall:       service.NewRecallService(messages, conversations, convCache),
		MarkRead:     service.NewMarkReadService(messages, conversations, unread, convCache),
		Conversation: service.NewConversationService(conversations, unread, convCache),
	}
}
