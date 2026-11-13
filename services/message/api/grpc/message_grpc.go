package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "isletalk/message/api/identity/v1"
	messagev1 "isletalk/message/api/message/v1"
	"isletalk/message/application/service"
	"isletalk/message/domain/entity"
	"isletalk/message/domain/valueobject"
)

// MessageGRPCServer 实现 message.v1.MessageService（4 RPC，T016/T021）。
type MessageGRPCServer struct {
	messagev1.UnimplementedMessageServiceServer
	sendMessage  *service.SendMessageService
	listMessages *service.ListMessagesService
	recall       *service.RecallService
	markRead     *service.MarkReadService
}

func NewMessageGRPCServer(
	sendMessage *service.SendMessageService,
	listMessages *service.ListMessagesService,
	recall *service.RecallService,
	markRead *service.MarkReadService,
) *MessageGRPCServer {
	return &MessageGRPCServer{
		sendMessage:  sendMessage,
		listMessages: listMessages,
		recall:       recall,
		markRead:     markRead,
	}
}

func (s *MessageGRPCServer) SendMessage(ctx context.Context, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	mid, err := valueobject.NewMessageId(req.GetMessageId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sender, err := valueobject.NewUserId(req.GetSenderId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	content, err := valueobject.NewMessageContent(req.GetContent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	typ, ok := messageTypeFromProto(req.GetType())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid message type")
	}
	m, err := s.sendMessage.Send(ctx, mid, req.GetConversationId(), sender, typ, content)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &messagev1.SendMessageResponse{Message: entityToProto(m)}, nil
}

func (s *MessageGRPCServer) ListMessages(ctx context.Context, req *messagev1.ListMessagesRequest) (*messagev1.ListMessagesResponse, error) {
	msgs, next, err := s.listMessages.List(ctx, req.GetConversationId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, grpcErr(err)
	}
	out := make([]*messagev1.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, entityToProto(m))
	}
	return &messagev1.ListMessagesResponse{Messages: out, NextPageToken: next}, nil
}

func (s *MessageGRPCServer) RecallMessage(ctx context.Context, req *messagev1.RecallMessageRequest) (*messagev1.RecallMessageResponse, error) {
	mid, err := valueobject.NewMessageId(req.GetMessageId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	sender, err := valueobject.NewUserId(req.GetSenderId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	recalled, err := s.recall.Recall(ctx, mid, sender)
	if err != nil {
		return nil, grpcErr(err)
	}
	return &messagev1.RecallMessageResponse{Recalled: recalled}, nil
}

func (s *MessageGRPCServer) MarkRead(ctx context.Context, req *messagev1.MarkReadRequest) (*messagev1.MarkReadResponse, error) {
	uid, err := valueobject.NewUserId(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.markRead.MarkRead(ctx, req.GetConversationId(), uid, req.GetLastReadSeq()); err != nil {
		return nil, grpcErr(err)
	}
	return &messagev1.MarkReadResponse{}, nil
}

// messageTypeFromProto 映射 pb 枚举到领域枚举。
func messageTypeFromProto(t messagev1.MessageType) (entity.MessageType, bool) {
	switch t {
	case messagev1.MessageType_MESSAGE_TYPE_TEXT:
		return entity.MessageText, true
	case messagev1.MessageType_MESSAGE_TYPE_IMAGE:
		return entity.MessageImage, true
	case messagev1.MessageType_MESSAGE_TYPE_FILE:
		return entity.MessageFile, true
	default:
		return "", false
	}
}

// messageTypeToProto 映射领域枚举到 pb 枚举。
func messageTypeToProto(t entity.MessageType) messagev1.MessageType {
	switch t {
	case entity.MessageImage:
		return messagev1.MessageType_MESSAGE_TYPE_IMAGE
	case entity.MessageFile:
		return messagev1.MessageType_MESSAGE_TYPE_FILE
	default:
		return messagev1.MessageType_MESSAGE_TYPE_TEXT
	}
}

// messageStatusToProto 映射领域状态到 pb 枚举。
func messageStatusToProto(st entity.MessageStatus) messagev1.MessageStatus {
	switch st {
	case entity.MessageDelivered:
		return messagev1.MessageStatus_MESSAGE_STATUS_DELIVERED
	case entity.MessageRead:
		return messagev1.MessageStatus_MESSAGE_STATUS_READ
	case entity.MessageFailed:
		return messagev1.MessageStatus_MESSAGE_STATUS_FAILED
	default:
		return messagev1.MessageStatus_MESSAGE_STATUS_SENT
	}
}

func entityToProto(m entity.Message) *messagev1.Message {
	return &messagev1.Message{
		MessageId:      m.MessageID().String(),
		ConversationId: m.ConversationID(),
		SenderId:       &identityv1.UserId{Value: m.SenderID().String()},
		Type:           messageTypeToProto(m.Type()),
		Content:        m.Content().String(),
		Seq:            m.Seq(),
		SentAt:         m.SentAt().UnixMilli(),
		Status:         messageStatusToProto(m.Status()),
		Recalled:       m.Recalled(),
	}
}

// grpcErr 映射用例错误到 gRPC status code（错误码注册表见 002）。
func grpcErr(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidMessageType),
		errors.Is(err, service.ErrInvalidPageToken),
		errors.Is(err, service.ErrInvalidConversationID),
		errors.Is(err, service.ErrNotParticipant),
		errors.Is(err, service.ErrNotSender):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFriends),
		errors.Is(err, service.ErrBlocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrMessageNotFound),
		errors.Is(err, service.ErrConversationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrRecallWindowExpired):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
