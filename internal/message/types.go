package message

import "time"

type Message struct {
	ID		  string	`json:"id"`
	ChatID	  string	`json:"chat_id"`
	SenderID  string	`json:"sender_id"`
	Seq		  int64		`json:"seq"`
	Content	  string	`json:"content"`
	CreatedAt time.Time	`json:"created_at"`
}

type IncomingMessage struct {
	Type	string `json:"type"`
	ChatID	string `json:"chat_id"`
	Content	string `json:"content"`
}

type OutgoingEvent struct {
	Type	string  `json:"type"`
	Message	Message `json:"message"`
}

type TypingEvent struct {
	Type   string `json:"type"`
	ChatID string `json:"chat_id"`
	UserID string `json:"user_id"`
}

type ReadReceipt struct {
	UserID		string `json:"user_id"`
	LastReadSeq int64  `json:"last_read_seq"`
}

type MarkAsReadRequest struct {
	UpToSeq int64 `json:"up_to_seq"`
}

type ReadEvent struct {
	Type	string `json:"type"`
	ChatID	string `json:"chat_id"`
	UserID	string `json:"user_id"`
	UpToSeq int64  `json:"up_to_seq"`
}
