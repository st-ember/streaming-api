package video

import "time"

type Video struct {
	ID          string
	Title       string
	Description string
	Duration    time.Duration
	Filename    string
	ResourceID  string
	Status      VideoStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewVideo(id, title, description, filename, resourceID string) (*Video, error) {
	if id == "" {
		return nil, ErrVideoIDEmpty
	}

	if filename == "" {
		return nil, ErrFilenameEmpty
	}

	if resourceID == "" {
		return nil, ErrResourceIDEmpty
	}

	return &Video{
		ID:          id,
		Title:       title,
		Description: description,
		Filename:    filename,
		ResourceID:  resourceID,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

// Update status
func (v *Video) MarkAsProcessing() error {
	if !v.CanBeProcessed() {
		return ErrCannotBeMarkedAsProcessing
	}

	v.Status = StatusProcessing
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func (v *Video) MarkAsFailed() error {
	if !v.IsProcessing() {
		return ErrCannotBeMarkedAsFailed
	}

	v.Status = StatusFailed
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func (v *Video) Publish() error {
	if !v.IsProcessing() {
		return ErrCannotBePublished
	}

	v.Status = StatusPublished
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func (v *Video) Archive() error {
	if !v.IsPublished() {
		return ErrCannotBeArchived
	}

	v.Status = StatusArchived
	v.UpdatedAt = time.Now().UTC()

	return nil
}

// Update fields
func (v *Video) UpdateTitle(title string) error {
	if title == "" {
		return ErrTitleEmpty
	}

	v.Title = title
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func (v *Video) UpdateDescription(description string) error {
	if description == "" {
		return ErrDescriptionEmpty
	}

	v.Description = description
	v.UpdatedAt = time.Now().UTC()

	return nil
}

func (v *Video) UpdateDuration(duration time.Duration) error {
	if v.Duration != 0 {
		return ErrDurationAlreadySet
	}

	if duration < 0 {
		return ErrDurationNegative
	}

	v.Duration = duration
	v.UpdatedAt = time.Now().UTC()

	return nil
}

// Status access
func (v *Video) IsPending() bool {
	return v.Status == StatusPending
}

func (v *Video) IsProcessing() bool {
	return v.Status == StatusProcessing
}

func (v *Video) IsPublished() bool {
	return v.Status == StatusPublished
}

func (v *Video) IsFailed() bool {
	return v.Status == StatusFailed
}

func (v *Video) IsArchived() bool {
	return v.Status == StatusArchived
}

func (v *Video) CanBeProcessed() bool {
	return v.Status == StatusPending || v.Status == StatusFailed
}
