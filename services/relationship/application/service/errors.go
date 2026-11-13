package service

import "errors"

// 用例错误（对外统一错误码，避免泄露关系状态）。
var (
	ErrSelfRequest         = errors.New("cannot request self")
	ErrStrangerAddDisabled = errors.New("stranger add disabled")
	ErrRequestExists       = errors.New("friend request already exists")
	ErrRequestNotFound     = errors.New("friend request not found")
	ErrAlreadyFriends      = errors.New("already friends")
	ErrBlocked             = errors.New("blocked")
	ErrRateLimited         = errors.New("rate limited")
	ErrRequestExpired      = errors.New("friend request expired")
)
