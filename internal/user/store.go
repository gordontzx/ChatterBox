package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	FindByEmail(ctx context.Context, email string) (*UserSummary, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) FindByEmail(ctx context.Context, email string) (*UserSummary, error) {
	var u UserSummary
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, email FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &u, err
}
