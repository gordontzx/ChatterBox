package auth

import "errors"

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials	  = errors.New("invalid email or password")
	ErrPasswordTooShort		  = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong		  = errors.New("password must be at most 72 characters")
	ErrInvalidEmail			  = errors.New("invalid email format")
)
