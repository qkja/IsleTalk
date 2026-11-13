package entity

import (
	"time"

	"isletalk/relationship/domain/valueobject"
)

// FriendRequestStatus 是好友申请状态（FR-002/003，pending → accepted/rejected）。
type FriendRequestStatus string

const (
	RequestPending  FriendRequestStatus = "pending"
	RequestAccepted FriendRequestStatus = "accepted"
	RequestRejected FriendRequestStatus = "rejected"
)

// FriendRequest 是好友申请聚合根。
type FriendRequest struct {
	initiatorID    valueobject.UserId
	targetID       valueobject.UserId
	requestMessage string
	status         FriendRequestStatus
	createdAt      time.Time
	respondedAt    time.Time
}

func NewFriendRequest(initiatorID, targetID valueobject.UserId, message string) FriendRequest {
	return FriendRequest{
		initiatorID:    initiatorID,
		targetID:       targetID,
		requestMessage: message,
		status:         RequestPending,
		createdAt:      time.Now().UTC(),
	}
}

func (r FriendRequest) InitiatorID() valueobject.UserId { return r.initiatorID }
func (r FriendRequest) TargetID() valueobject.UserId    { return r.targetID }
func (r FriendRequest) RequestMessage() string          { return r.requestMessage }
func (r FriendRequest) Status() FriendRequestStatus     { return r.status }
func (r FriendRequest) CreatedAt() time.Time            { return r.createdAt }

func (r FriendRequest) IsPending() bool { return r.status == RequestPending }

func (r *FriendRequest) Accept() { r.status = RequestAccepted; r.respondedAt = time.Now().UTC() }
func (r *FriendRequest) Reject() { r.status = RequestRejected; r.respondedAt = time.Now().UTC() }
