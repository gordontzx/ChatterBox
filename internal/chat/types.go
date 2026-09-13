package chat

import "time"

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type Member struct {
	ID	 string	`json:"id"`
	Name string	`json:"name"`
}

type Chat struct {
	ID		  string	`json:"id"`
	Type	  ChatType  `json:"type"`
	Name	  *string	`json:"name,omitempty"`	// nil for direct chats
	CreatedAt time.Time	`json:"created_at"`
	Members   []Member	`json:"members"`
}

type CreateDirectChatRequest struct {
	OtherUserID string `json:"other_user_id"`
}

type CreateGroupChatRequest struct {
	Name	  string   `json:"name"`
	MemberIDs []string `json:"member_ids"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

type ChatEvent struct {
	Type   string  `json:"type"` // "chat_created", "member_added", "member_left"
	Chat   *Chat   `json:"chat"`
	Member *Member `json:"member,omitempty"` // the member who was added/left
}
