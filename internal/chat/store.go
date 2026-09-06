package chat

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	FindDirectChat(ctx context.Context, user1, user2 string) (*Chat, error)
	CreateDirectChat(ctx context.Context, user1, user2 string) (*Chat, error)
	ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// Looks for existing chat with exactly user1 and user2. Returns (nil, nil) if none exists
func (s *PostgresStore) FindDirectChat(ctx context.Context, user1, user2 string) (*Chat, error) {
	var c Chat
	err := s.pool.QueryRow(ctx,
		`SELECT c.id, c.created_at
		 FROM chats c
		 WHERE c.id IN (
			 SELECT chat_id FROM chat_members WHERE user_id = $1
			 INTERSECT
			 SELECT chat_id FROM chat_members WHERE user_id = $2
		 )
		 AND (SELECT COUNT(*) FROM chat_members WHERE chat_id = c.id) = 2`,
		user1, user2,
	).Scan(&c.ID, &c.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	c.MemberIDs = []string{user1, user2}
	return &c, nil
}

func (s *PostgresStore) CreateDirectChat(ctx context.Context, user1, user2 string) (*Chat, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var c Chat
	err = tx.QueryRow(ctx,
		`INSERT into chats DEFAULT VALUES RETURNING id, created_at`,
	).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT into chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)`,
		c.ID, user1, user2,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	c.MemberIDs = []string{user1, user2}
	return &c, nil
}

// Returns all chats the given user is a member of
func (s *PostgresStore) ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT c.id, c.created_at
		 FROM chats c
		 JOIN chat_members cm ON cm.chat_id = c.id
		 WHERE cm.user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		var c Chat
		if err := rows.Scan(&c.ID, &c.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &c)
	}
	return chats, rows.Err()
}

func (s *PostgresStore) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		userID,
	).Scan(&exists)
	return exists, err
}
