package chat

import "errors"

var (
	ErrCannotChatWithSelf = errors.New("cannot create a chat with yourself")
	ErrUserNotFound		  = errors.New("user not found")
)
