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
	CreateGroupChat(ctx context.Context, name string, memberIDs []string) (*Chat, error)
	AddMember(ctx context.Context, chatID, userID string) error
	RemoveMember(ctx context.Context, chatID, userID string) error
	GetChat(ctx context.Context, chatID string) (*Chat, error)
	ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error)
	IsChatMember(ctx context.Context, chatID, userID string) (bool, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// Looks for existing direct chat with exactly user1 and user2. Returns (nil, nil) if none exists
func (s *PostgresStore) FindDirectChat(ctx context.Context, user1, user2 string) (*Chat, error) {
	var chatID string
	err := s.pool.QueryRow(ctx,
		`SELECT c.id
		 FROM chats c
		 WHERE c.chat_type = 'direct'
		 AND c.id IN (
			 SELECT chat_id FROM chat_members WHERE user_id = $1
			 INTERSECT
			 SELECT chat_id FROM chat_members WHERE user_id = $2
		 )`,
		user1, user2,
	).Scan(&chatID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return s.GetChat(ctx, chatID)
}

func (s *PostgresStore) CreateDirectChat(ctx context.Context, user1, user2 string) (*Chat, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var chatID string
	err = tx.QueryRow(ctx,
		`INSERT into chats (chat_type) VALUES ('direct') RETURNING id`,
	).Scan(&chatID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT into chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)`,
		chatID, user1, user2,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetChat(ctx, chatID)
}

func (s *PostgresStore) CreateGroupChat(ctx context.Context, name string, memberIDs []string) (*Chat, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var chatID string
	err = tx.QueryRow(ctx,
		`INSERT INTO chats (chat_type, name) VALUES ('group', $1) RETURNING id`,
		name,
	).Scan(&chatID)
	if err != nil {
		return nil, err
	}

	batch := &pgx.Batch{}
	for _, userID := range memberIDs {
		batch.Queue(`INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)`, chatID, userID)
	}
	br := tx.SendBatch(ctx, batch)
	for range memberIDs {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return nil, err
		}
	}
	if err := br.Close(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetChat(ctx, chatID)
}

func (s *PostgresStore) AddMember(ctx context.Context, chatID, userID string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT into chat_members (chat_id, user_id) VALUES ($1, $2)
		 ON CONFLICT (chat_id, user_id) DO NOTHING`,
		chatID, userID,
	)
	return err
}

func (s *PostgresStore) RemoveMember(ctx context.Context, chatID, userID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM chat_members
		 WHERE chat_id = $1 AND user_id = $2`,
		chatID, userID,
	)
	return err
}

func (s *PostgresStore) IsChatMember(ctx context.Context, chatID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2)`,
		chatID, userID,
	).Scan(&exists)
	return exists, err
}

func (s *PostgresStore) GetChat(ctx context.Context, chatID string) (*Chat, error) {
	var c Chat
	var chatType string
	err := s.pool.QueryRow(ctx,
		`SELECT id, chat_type, name, created_at
		 FROM chats
		 WHERE id = $1`,
		chatID,
	).Scan(&c.ID, &chatType, &c.Name, &c.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Type = ChatType(chatType)

	members, err := s.getChatMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}
	c.Members = members

	return &c, nil
}

// Gets members info (id + name) for a single chat
func (s *PostgresStore) getChatMembers(ctx context.Context, chatID string) ([]Member, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT u.id, u.name
		 FROM chat_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.chat_id = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

// Returns all chats the given user is a member of
func (s *PostgresStore) ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT c.id, c.chat_type, c.name, c.created_at
		 FROM chats c
		 JOIN chat_members cm ON cm.chat_id = c.id
		 WHERE cm.user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}

	var chats []*Chat
	chatsByID := make(map[string]*Chat)
	var chatIDs []string

	for rows.Next() {
		var c Chat
		var chatType string
		if err := rows.Scan(&c.ID, &chatType, &c.Name, &c.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		c.Type = ChatType(chatType)
		chats = append(chats, &c)
		chatsByID[c.ID] = &c
		chatIDs = append(chatIDs, c.ID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(chatIDs) == 0 {
		return chats, nil
	}

	// Get all members from found chat IDs
	memberRows, err := s.pool.Query(ctx,
		`SELECT cm.chat_id, u.id, u.name
		 FROM chat_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.chat_id = ANY($1)
		 ORDER BY u.name`,
		chatIDs,
	)
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var chatID string
		var m Member
		if err := memberRows.Scan(&chatID, &m.ID, &m.Name); err != nil {
			return nil, err
		}
		if c, ok := chatsByID[chatID]; ok {
			c.Members = append(c.Members, m)
		}
	}

	return chats, memberRows.Err()
}

func (s *PostgresStore) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		userID,
	).Scan(&exists)
	return exists, err
}
