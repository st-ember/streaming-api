package job

import "time"

type Job struct {
	ID        string
	VideoID   string
	Type      JobType
	Status    JobStatus
	Result    string
	ErrorMsg  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewJob(id, videoID string, jobType JobType) (*Job, error) {
	if id == "" {
		return nil, ErrJobIDEmpty
	}

	if videoID == "" {
		return nil, ErrVideoIDEmpty
	}

	if !jobType.IsValid() {
		return nil, ErrJobTypeInvalid
	}

	return &Job{
		ID:        id,
		VideoID:   videoID,
		Type:      jobType,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// Lifycycle management
func (j *Job) Start() error {
	if !j.CanBeStarted() {
		return ErrCannotBeStarted
	}

	j.Status = StatusRunning
	j.UpdatedAt = time.Now().UTC()

	return nil
}

func (j *Job) Complete(result string) error {
	if !j.IsRunning() {
		return ErrCannotBeCompleted
	}

	j.Status = StatusCompleted
	j.Result = result
	j.UpdatedAt = time.Now().UTC()

	return nil
}

func (j *Job) MarkAsFailed(errMsg string) error {
	if !j.IsRunning() {
		return ErrCannotBeMarkedAsFailed
	}

	j.Status = StatusFailed
	j.ErrorMsg = errMsg
	j.UpdatedAt = time.Now().UTC()

	return nil
}

// Status access
func (j *Job) IsPending() bool {
	return j.Status == StatusPending
}

func (j *Job) IsRunning() bool {
	return j.Status == StatusRunning
}

func (j *Job) IsCompleted() bool {
	return j.Status == StatusCompleted
}

func (j *Job) IsFailed() bool {
	return j.Status == StatusFailed
}

func (j *Job) CanBeStarted() bool {
	return j.Status == StatusPending || j.Status == StatusFailed
}
