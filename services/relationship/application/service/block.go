package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// BlockService 是拉黑/解除拉黑用例（FR-008，单向）。
type BlockService struct {
	blocks   port.BlockRepository
	requests port.FriendRequestRepository
}

func NewBlockService(blocks port.BlockRepository, requests port.FriendRequestRepository) *BlockService {
	return &BlockService{blocks: blocks, requests: requests}
}

// SetBlock 拉黑（blocked=true）或解除拉黑（blocked=false）。
// 拉黑时使对方发来的待处理申请失效（Edge Case）。
func (s *BlockService) SetBlock(ctx context.Context, ownerID, targetID valueobject.UserId, blocked bool) error {
	slog.Info("set block", "owner", ownerID.String(), "target", targetID.String(), "blocked", blocked)
	if err := s.blocks.Set(ctx, ownerID, targetID, blocked); err != nil {
		slog.Error("set block failed", "err", err, "owner", ownerID.String(), "target", targetID.String())
		return err
	}
	if blocked {
		// 使对方发来的待处理申请失效（对方 target → 我 owner 的申请）
		if req, err := s.requests.Find(ctx, targetID, ownerID); err == nil && req.IsPending() {
			req.Reject()
			if err := s.requests.Update(ctx, req); err != nil {
				slog.Error("block reject pending request failed", "err", err)
				return err
			}
			slog.Info("block rejected pending request", "initiator", targetID.String(), "target", ownerID.String())
		}
	}
	slog.Info("set block success", "owner", ownerID.String(), "target", targetID.String())
	return nil
}
