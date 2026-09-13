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

func (s *Service) CreateOrGetDirectChat(ctx context.Context, callerID, otherUserID string) (*Chat, bool, error) {
	if callerID == otherUserID {
		return nil, false, ErrCannotChatWithSelf
	}

	exists, err := s.store.UserExists(ctx, otherUserID)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, ErrUserNotFound
	}

	existing, err := s.store.FindDirectChat(ctx, callerID, otherUserID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	created, err := s.store.CreateDirectChat(ctx, callerID, otherUserID)
	if err != nil {
		return nil, false, err
	}

	return created, true, nil
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

func (s *Service) GetChat(ctx context.Context, callerID, chatID string) (*Chat, error) {
	chat, err := s.store.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat == nil {
		return nil, ErrChatNotFound
	}

	isMember, err := s.store.IsChatMember(ctx, chatID, callerID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotChatMember
	}

	return chat, nil
}

func (s *Service) AddMember(ctx context.Context, callerID, chatID, userID string) error {
	chat, err := s.GetChat(ctx, callerID, chatID)
	if err != nil {
		return err
	}
	if chat.Type != ChatTypeGroup {
		return ErrNotGroupChat
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

func (s *Service) RemoveMember(ctx context.Context, callerID, chatID, targetUserID string) error {
	chat, err := s.GetChat(ctx, callerID, chatID)
	if err != nil {
		return err
	}
	if chat.Type != ChatTypeGroup {
		return ErrNotGroupChat
	}

	exists, err := s.store.UserExists(ctx, targetUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	return s.store.RemoveMember(ctx, chatID, targetUserID)
}

func (s *Service) LeaveChat(ctx context.Context, callerID, chatID string) error {
	return s.RemoveMember(ctx, callerID, chatID, callerID)
}

func (s *Service) ListContactIDs(ctx context.Context, userID string) ([]string, error) {
	chats, err := s.store.ListChatsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := map[string]struct{}{userID: {}}
	var contacts []string
	for _, c := range chats {
		for _, m := range c.Members {
			if _, exists := seen[m.ID]; !exists {
				seen[m.ID] = struct{}{}
				contacts = append(contacts, m.ID)
			}
		}
	}

	return contacts, nil
}

func (s *Service) ShareAnyChat(ctx context.Context, user1, user2 string) (bool, error) {
	return s.store.ShareAnyChat(ctx, user1, user2)
}
