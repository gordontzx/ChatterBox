package presence

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	UpdateLastSeen(ctx context.Context, userID string, at time.Time) error
	GetLastSeen(ctx context.Context, userID string) (*time.Time, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) UpdateLastSeen(ctx context.Context, userID string, at time.Time) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET last_seen_at = $1 WHERE id = $2`,
		at, userID,
	)
	return err
}

func (s *PostgresStore) GetLastSeen(ctx context.Context, userID string) (*time.Time, error) {
	var lastSeen *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT last_seen_at FROM users WHERE id = $1`,
		userID,
	).Scan(&lastSeen)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return lastSeen, err
}
