package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	FindByEmail(ctx context.Context, email string) (*PublicUser, error)
	FindByID(ctx context.Context, userID string) (*PublicUser, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) FindByEmail(ctx context.Context, email string) (*PublicUser, error) {
	var u PublicUser
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, email FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (s *PostgresStore) FindByID(ctx context.Context, userID string) (*PublicUser, error) {
	var u PublicUser
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, email FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Name, &u.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}
