package video

import "errors"

var (
	ErrVideoIDEmpty               = errors.New("video id cannot be empty")
	ErrResourceIDEmpty            = errors.New("resource id cannot be empty")
	ErrCannotBeMarkedAsProcessing = errors.New("video cannot be marked as processing")
	ErrCannotBeMarkedAsFailed     = errors.New("video cannot be marked as failed")
	ErrCannotBePublished          = errors.New("video cannot be published")
	ErrCannotBeArchived           = errors.New("video cannot be archived")
	ErrTitleEmpty                 = errors.New("video title cannot be empty")
	ErrDescriptionEmpty           = errors.New("video description cannot be empty")
	ErrDurationAlreadySet         = errors.New("video duration has already been set")
	ErrDurationNegative           = errors.New("video duration cannot be negative")
)
