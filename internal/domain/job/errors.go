package job

import "errors"

var (
	ErrJobIDEmpty             = errors.New("job id cannot be empty")
	ErrVideoIDEmpty           = errors.New("video id cannot be empty")
	ErrJobTypeInvalid         = errors.New("job type is invalid")
	ErrCannotBeStarted        = errors.New("job cannot be started")
	ErrCannotBeCompleted      = errors.New("job cannot be completed")
	ErrCannotBeMarkedAsFailed = errors.New("job cannot be marked as failed")
)
