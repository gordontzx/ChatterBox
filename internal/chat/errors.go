package chat

import "errors"

var (
	ErrCannotChatWithSelf = errors.New("cannot create a chat with yourself")
	ErrUserNotFound		  = errors.New("user not found")
	ErrGroupNameRequired  = errors.New("group chat name is required")
	ErrNotGroupChat		  = errors.New("this operation only applies to group chats")
	ErrNotChatMember	  = errors.New("caller is not a member of this chat")
	ErrAlreadyMember	  = errors.New("user is already a member of this chat")
)
