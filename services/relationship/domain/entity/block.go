package entity

import (
	"time"

	"isletalk/relationship/domain/valueobject"
)

// BlockStatus 是拉黑状态（FR-008，blocked ↔ unblocked，单向）。
type BlockStatus string

const (
	BlockActive   BlockStatus = "blocked"
	BlockInactive BlockStatus = "unblocked"
)

// Block 是拉黑关系聚合根（单向，owner → target）。
type Block struct {
	ownerID   valueobject.UserId
	targetID  valueobject.UserId
	status    BlockStatus
	createdAt time.Time
}

func NewBlock(ownerID, targetID valueobject.UserId) Block {
	return Block{ownerID: ownerID, targetID: targetID, status: BlockActive, createdAt: time.Now().UTC()}
}

func (b Block) OwnerID() valueobject.UserId  { return b.ownerID }
func (b Block) TargetID() valueobject.UserId { return b.targetID }
func (b Block) Status() BlockStatus          { return b.status }

func (b Block) IsActive() bool { return b.status == BlockActive }

func (b *Block) Unblock() { b.status = BlockInactive }
