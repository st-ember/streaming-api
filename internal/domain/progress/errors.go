package progress

import "errors"

var (
	ErrTotalFramesZero       = errors.New("total frames cannot be zero")
	ErrCannotBeUpdated       = errors.New("current frames cannot be updated")
	ErrCannotBeMarkedAsEnd   = errors.New("progress cannot be marked as ended")
	ErrCannotBeMarkedAsError = errors.New("progress cannot be marked as error")
)
