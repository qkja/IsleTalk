package entity

import (
	"time"

	"isletalk/relationship/domain/valueobject"
)

// FriendshipStatus 是好友关系状态（FR-007，data-model：accepted → deleted）。
type FriendshipStatus string

const (
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipDeleted  FriendshipStatus = "deleted"
)

// Friendship 是好友关系聚合根。一对好友 = 两条文档（A→B 与 B→A，SC-002 双向一致）。
type Friendship struct {
	userID    valueobject.UserId
	friendID  valueobject.UserId
	createdAt time.Time
	status    FriendshipStatus
}

func NewFriendship(userID, friendID valueobject.UserId) Friendship {
	return Friendship{
		userID:    userID,
		friendID:  friendID,
		createdAt: time.Now().UTC(),
		status:    FriendshipAccepted,
	}
}

func (f Friendship) UserID() valueobject.UserId   { return f.userID }
func (f Friendship) FriendID() valueobject.UserId { return f.friendID }
func (f Friendship) CreatedAt() time.Time         { return f.createdAt }
func (f Friendship) Status() FriendshipStatus     { return f.status }

func (f Friendship) IsActive() bool { return f.status == FriendshipAccepted }

// Delete 解除关系（双向解除时两条都删）。
func (f *Friendship) Delete() { f.status = FriendshipDeleted }
