package chat

import "time"

type Chat struct {
	ID		  string
	CreatedAt time.Time
	MemberIDs []string
}

type CreateDirectChatRequest struct {
	OtherUserID string `json:"other_user_id"`
}
