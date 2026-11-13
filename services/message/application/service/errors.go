package service

import "errors"

// 用例错误（对外统一错误码，映射见 api 层与 002 error-registry）。
var (
	ErrNotFriends            = errors.New("not friends")
	ErrBlocked               = errors.New("blocked")
	ErrMessageNotFound       = errors.New("message not found")
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrRecallWindowExpired   = errors.New("recall window expired")
	ErrNotSender             = errors.New("not the message sender")
	ErrInvalidMessageType    = errors.New("invalid message type")
	ErrInvalidPageToken      = errors.New("invalid page token")
	ErrNotParticipant        = errors.New("not a conversation participant")
	ErrInvalidConversationID = errors.New("invalid conversation id")
)
