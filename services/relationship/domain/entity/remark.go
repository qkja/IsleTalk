package entity

import (
	"time"

	"isletalk/relationship/domain/valueobject"
)

// Remark 是好友备注聚合根（FR-005，单向，仅 owner 可见）。
type Remark struct {
	ownerID   valueobject.UserId
	friendID  valueobject.UserId
	content   valueobject.RemarkText
	updatedAt time.Time
}

func NewRemark(ownerID, friendID valueobject.UserId, content valueobject.RemarkText) Remark {
	return Remark{ownerID: ownerID, friendID: friendID, content: content, updatedAt: time.Now().UTC()}
}

func (r Remark) OwnerID() valueobject.UserId     { return r.ownerID }
func (r Remark) FriendID() valueobject.UserId    { return r.friendID }
func (r Remark) Content() valueobject.RemarkText { return r.content }

// Update 更新备注内容。
func (r *Remark) Update(content valueobject.RemarkText) {
	r.content = content
	r.updatedAt = time.Now().UTC()
}
