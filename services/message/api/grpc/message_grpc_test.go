package grpc

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	messagev1 "isletalk/message/api/message/v1"
	"isletalk/message/application/service"
	"isletalk/message/domain/entity"
)

func TestMessageTypeFromProto(t *testing.T) {
	cases := []struct {
		pb messagev1.MessageType
		dm entity.MessageType
		ok bool
	}{
		{messagev1.MessageType_MESSAGE_TYPE_TEXT, entity.MessageText, true},
		{messagev1.MessageType_MESSAGE_TYPE_IMAGE, entity.MessageImage, true},
		{messagev1.MessageType_MESSAGE_TYPE_FILE, entity.MessageFile, true},
		{messagev1.MessageType_MESSAGE_TYPE_UNSPECIFIED, "", false},
	}
	for _, tc := range cases {
		got, ok := messageTypeFromProto(tc.pb)
		if ok != tc.ok || got != tc.dm {
			t.Fatalf("messageTypeFromProto(%v) = %v, %v; want %v, %v", tc.pb, got, ok, tc.dm, tc.ok)
		}
	}
}

func TestMessageTypeToProto(t *testing.T) {
	if messageTypeToProto(entity.MessageText) != messagev1.MessageType_MESSAGE_TYPE_TEXT ||
		messageTypeToProto(entity.MessageImage) != messagev1.MessageType_MESSAGE_TYPE_IMAGE ||
		messageTypeToProto(entity.MessageFile) != messagev1.MessageType_MESSAGE_TYPE_FILE {
		t.Fatal("messageTypeToProto mapping wrong")
	}
}

func TestMessageStatusToProto(t *testing.T) {
	cases := map[entity.MessageStatus]messagev1.MessageStatus{
		entity.MessageSent:      messagev1.MessageStatus_MESSAGE_STATUS_SENT,
		entity.MessageDelivered: messagev1.MessageStatus_MESSAGE_STATUS_DELIVERED,
		entity.MessageRead:      messagev1.MessageStatus_MESSAGE_STATUS_READ,
		entity.MessageFailed:    messagev1.MessageStatus_MESSAGE_STATUS_FAILED,
	}
	for dm, pb := range cases {
		if messageStatusToProto(dm) != pb {
			t.Fatalf("messageStatusToProto(%v) wrong", dm)
		}
	}
}

func TestGRPCErrMapping(t *testing.T) {
	cases := []struct {
		err  error
		code codes.Code
	}{
		{service.ErrNotFriends, codes.PermissionDenied},
		{service.ErrBlocked, codes.PermissionDenied},
		{service.ErrMessageNotFound, codes.NotFound},
		{service.ErrConversationNotFound, codes.NotFound},
		{service.ErrRecallWindowExpired, codes.FailedPrecondition},
		{service.ErrInvalidMessageType, codes.InvalidArgument},
		{service.ErrInvalidPageToken, codes.InvalidArgument},
		{service.ErrNotSender, codes.InvalidArgument},
		{service.ErrNotParticipant, codes.InvalidArgument},
		{errors.New("unknown"), codes.Internal},
	}
	for _, tc := range cases {
		st, ok := status.FromError(grpcErr(tc.err))
		if !ok || st.Code() != tc.code {
			t.Fatalf("grpcErr(%v) = %v, want code %v", tc.err, st.Code(), tc.code)
		}
	}
}
