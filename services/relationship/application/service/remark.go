package service

import (
	"context"
	"log/slog"

	"isletalk/relationship/domain/entity"
	"isletalk/relationship/domain/port"
	"isletalk/relationship/domain/valueobject"
)

// RemarkService 是好友备注用例（FR-005，单向，仅设置方可见）。
type RemarkService struct {
	remarks port.RemarkRepository
}

func NewRemarkService(remarks port.RemarkRepository) *RemarkService {
	return &RemarkService{remarks: remarks}
}

// SetRemark 设置/修改好友备注。
func (s *RemarkService) SetRemark(ctx context.Context, ownerID, friendID valueobject.UserId, content string) (entity.Remark, error) {
	slog.Info("set remark", "owner", ownerID.String(), "friend", friendID.String())
	txt, err := valueobject.NewRemarkText(content)
	if err != nil {
		slog.Error("set remark invalid text", "err", err)
		return entity.Remark{}, err
	}
	remark := entity.NewRemark(ownerID, friendID, txt)
	saved, err := s.remarks.Save(ctx, remark)
	if err != nil {
		slog.Error("set remark save failed", "err", err)
		return entity.Remark{}, err
	}
	slog.Info("set remark success", "owner", ownerID.String(), "friend", friendID.String())
	return saved, nil
}
