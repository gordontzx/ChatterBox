package chat

import "context"

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
