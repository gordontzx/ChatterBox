package user

import (
	"context"
	"strings"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*PublicUser, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	u, err := s.store.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func (s *Service) FindByID(ctx context.Context, userID string) (*PublicUser, error) {
	u, err := s.store.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	return u, nil
}

func normalizeEmail(email string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(email))
	if trimmed == "" || !strings.Contains(trimmed, "@") || strings.ContainsAny(trimmed, " \t\n") {
		return "", ErrInvalidEmail
	}
	return trimmed, nil
}
