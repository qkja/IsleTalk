package di

import (
	"isletalk/relationship/application/service"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/infrastructure/cache"
)

// Container 是组合根（宪法原则 V：组合根 MUST 唯一且位于最外层）。
type Container struct {
	FriendRequest *service.FriendRequestService
	Respond       *service.RespondService
	Block         *service.BlockService
	Remark        *service.RemarkService
	FriendList    *service.FriendListService
	CanMessage    *service.CanMessageService
}

// NewContainer 装配全部 application 服务。
func NewContainer(
	friendships port.FriendshipRepository,
	requests port.FriendRequestRepository,
	blocks port.BlockRepository,
	remarks port.RemarkRepository,
	identity port.IdentityClient,
) *Container {
	return &Container{
		FriendRequest: service.NewFriendRequestService(requests, friendships, blocks, identity, cache.NewRateLimiter(nil, 0, 10)),
		Respond:       service.NewRespondService(requests, friendships),
		Block:         service.NewBlockService(blocks, requests),
		Remark:        service.NewRemarkService(remarks),
		FriendList:    service.NewFriendListService(friendships),
		CanMessage:    service.NewCanMessageService(friendships, blocks),
	}
}
