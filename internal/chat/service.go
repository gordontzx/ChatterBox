package chat

import (
	"context"
	"strings"
)

const minGroupOtherMembers = 1

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateOrGetDirectChat(ctx context.Context, callerID, otherUserID string) (*Chat, error) {
	if callerID == otherUserID {
		return nil, ErrCannotChatWithSelf
	}

	exists, err := s.store.UserExists(ctx, otherUserID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	existing, err := s.store.FindDirectChat(ctx, callerID, otherUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	return s.store.CreateDirectChat(ctx, callerID, otherUserID)
}

func (s *Service) ListChats(ctx context.Context, userID string) ([]*Chat, error) {
	return s.store.ListChatsForUser(ctx, userID)
}

func (s *Service) CreateGroupChat(ctx context.Context, callerID string, req CreateGroupChatRequest) (*Chat, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrGroupNameRequired
	}

	seen := map[string]bool{callerID: true}
	var others []string
	for _, id := range req.MemberIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		others = append(others, id)
	}

	if len(others) < minGroupOtherMembers {
		return nil, ErrGroupTooSmall
	}

	for _, id := range others {
		exists, err := s.store.UserExists(ctx, id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrUserNotFound
		}
	}

	allMembers := append([]string{callerID}, others...)
	return s.store.CreateGroupChat(ctx, name, allMembers)
}

func (s *Service) AddMember(ctx context.Context, callerID, chatID, userID string) error {
	chat, err := s.store.GetChat(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		return ErrChatNotFound
	}
	if chat.Type != ChatTypeGroup {
		return ErrNotGroupChat
	}

	isMember, err := s.store.IsChatMember(ctx, chatID, callerID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrAlreadyMember
	}

	exists, err := s.store.UserExists(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	alreadyMember, err := s.store.IsChatMember(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if alreadyMember {
		return ErrAlreadyMember
	}

	return s.store.AddMember(ctx, chatID, userID)
}
