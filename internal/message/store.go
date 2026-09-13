package message

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	IsChatMember(ctx context.Context, chatID, userID string) (bool, error)
	OtherChatMemberIDs(ctx context.Context, chatID, excludeUserID string) ([]string, error)
	CreateMessage(ctx context.Context, chatID, senderID, content string) (*Message, error)
	GetMessagesSince(ctx context.Context, chatID string, sinceSeq int64, limit int) ([]Message, error)
	UpsertReadReceipt(ctx context.Context, chatID, userID string, upToSeq int64) error
	GetReadReceiptsForChat(ctx context.Context, chatID string) ([]ReadReceipt, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) IsChatMember(ctx context.Context, chatID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2
		)`,
		chatID, userID,
	).Scan(&exists)
	return exists, err
}

func (s *PostgresStore) OtherChatMemberIDs(ctx context.Context, chatID, excludeUserID string) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT user_id FROM chat_members WHERE chat_id = $1 AND user_id != $2`,
		chatID, excludeUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *PostgresStore) CreateMessage(ctx context.Context, chatID, senderID, content string) (*Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var nextSeq int64
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(seq), 0) + 1 FROM messages WHERE chat_id = $1`,
		chatID,
	).Scan(&nextSeq)
	if err != nil {
		return nil, err
	}

	var m Message
	err = tx.QueryRow(ctx,
		`INSERT INTO messages (chat_id, sender_id, seq, content)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, chat_id, sender_id, seq, content, created_at`,
		chatID, senderID, nextSeq, content,
	).Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Seq, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &m, nil
}

// Return messages with seq > sinceSeq, capped at limit
func (s *PostgresStore) GetMessagesSince(ctx context.Context, chatID string, sinceSeq int64, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, chat_id, sender_id, seq, content, created_at
		 FROM messages
		 WHERE chat_id = $1 AND seq > $2
		 ORDER BY seq ASC
		 LIMIT $3`,
		chatID, sinceSeq, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Seq, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *PostgresStore) UpsertReadReceipt(ctx context.Context, chatID, userID string, upToSeq int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO read_receipts (chat_id, user_id, last_read_seq)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (chat_id, user_id)
		 DO UPDATE SET last_read_seq = GREATEST(read_receipts.last_read_seq, EXCLUDED.last_read_seq),
		 			   updated_at = now()`,
		chatID, userID, upToSeq,
	)
	return err
}

func (s *PostgresStore) GetReadReceiptsForChat(ctx context.Context, chatID string) ([]ReadReceipt, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT user_id, last_read_seq FROM read_receipts WHERE chat_id = $1`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []ReadReceipt
	for rows.Next() {
		var r ReadReceipt
		if err := rows.Scan(&r.UserID, &r.LastReadSeq); err != nil {
			return nil, err
		}
		receipts = append(receipts, r)
	}

	return receipts, rows.Err()
}
