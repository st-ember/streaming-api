package user

import "errors"

var (
	ErrIDEmpty       = errors.New("user id cannot be empty")
	ErrEmailEmpty    = errors.New("email cannot be empty")
	ErrUsernameEmpty = errors.New("username cannot be empty")
	ErrPwdHashEmpty  = errors.New("password hash cannot be empty")
)
