package chat

import "time"

type Member struct {
	ID	 string	`json:"id"`
	Name string	`json:"name"`
}

type Chat struct {
	ID		  string	`json:"id"`
	CreatedAt time.Time	`json:"created_at"`
	Members   []Member	`json:"members"`
}

type CreateDirectChatRequest struct {
	OtherUserID string `json:"other_user_id"`
}
