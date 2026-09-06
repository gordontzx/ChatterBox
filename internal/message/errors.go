package message

import "errors"

var (
	ErrNotChatMember  = errors.New("sender is not a member of this chat")
	ErrEmptyContent	  = errors.New("message cannot be empty")
	ErrContentTooLong = errors.New("message content exceeds maximum length")
)
