package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateSession(ctx context.Context, userID, deviceName, refreshTokenHash string) error
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// Inserts a new user and returns it
func (s *PostgresStore) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name`,
		email, passwordHash, name,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, err
	}

	return &u, nil
}

// Returns (nil, nil) if no user is found — this is not treated as an error
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name
		 FROM users
		 WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// Inserts a new session row for a logged-in user/device
func (s *PostgresStore) CreateSession(ctx context.Context, userID, deviceName, refreshTokenHash string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT into sessions (user_id, device_name, refresh_token_hash)
		 VALUES ($1, $2, $3)`,
		userID, deviceName, refreshTokenHash,
	)
	return err
}
