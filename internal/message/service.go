package message

import (
	"context"
	"strings"
)

const (
	maxContentLength = 4096
	defaultMessageLimit = 100
	maxMessageLimit = 200
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// bundles message with who else needs to receive it
type SendResult struct {
	Message		   Message
	OtherMemberIDs []string
}

func (s *Service) SendMessage(ctx context.Context, senderID, chatID, content string) (*SendResult, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > maxContentLength {
		return nil, ErrContentTooLong
	}

	isMember, err := s.store.IsChatMember(ctx, chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	msg, err := s.store.CreateMessage(ctx, chatID, senderID, content)
	if err != nil {
		return nil, err
	}

	otherIDs, err := s.store.OtherChatMemberIDs(ctx, chatID, senderID)
	if err != nil {
		return nil, err
	}

	return &SendResult{Message: *msg, OtherMemberIDs: otherIDs}, nil
}

func (s *Service) SendTypingIndicator(ctx context.Context, senderID, chatID string) ([]string, error) {
	isMember, err := s.store.IsChatMember(ctx, chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	return s.store.OtherChatMemberIDs(ctx, chatID, senderID)
}

func (s *Service) GetMessagesSince(ctx context.Context, callerID, chatID string, sinceSeq int64, limit int) ([]Message, error) {
	isMember, err := s.store.IsChatMember(ctx, chatID, callerID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	limit = clampLimit(limit)

	return s.store.GetMessagesSince(ctx, chatID, sinceSeq, limit)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultMessageLimit
	}
	return min(limit, maxMessageLimit)
}

// Records that caller has read up to the given seq in the chat and returns other members who should be notified
func (s *Service) MarkAsRead(ctx context.Context, userID, chatID string, upToSeq int64) ([]string, error) {
	if upToSeq < 0 {
		return nil, ErrInvalidSeq
	}

	isMember, err := s.store.IsChatMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	if err := s.store.UpsertReadReceipt(ctx, chatID, userID, upToSeq); err != nil {
		return nil, err
	}

	return s.store.OtherChatMemberIDs(ctx, chatID, userID)
}

func (s *Service) GetReadReceipts(ctx context.Context, callerID, chatID string) ([]ReadReceipt, error) {
	isMember, err := s.store.IsChatMember(ctx, chatID, callerID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	return s.store.GetReadReceiptsForChat(ctx, chatID)
}
