package user

import (
	"context"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*UserSummary, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}

	return s.store.FindByEmail(ctx, email)
}
