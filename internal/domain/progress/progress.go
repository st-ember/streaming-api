package progress

// Progress represents the transcoding progress of a video.
// It tracks the number of processed frames relative to the total.
type Progress struct {
	TotalFrames   int64
	CurrentFrames int64
	Status        ProgressStatus
	Percentage    int
}

// NewProgress creates a new Progress instance with the specified total frames.
// Returns an error if totalFrames is zero.
func NewProgress(totalFrames int64) (*Progress, error) {
	if totalFrames == 0 {
		return nil, ErrTotalFramesZero
	}

	return &Progress{
		TotalFrames: totalFrames,
		Status:      StatusContinue,
	}, nil
}

// UpdateCurrentFrames updates the number of processed frames and recalculates the percentage.
// It caps the percentage at 100 to handle potential inaccuracies in FFmpeg's frame reporting.
func (p *Progress) UpdateCurrentFrames(currentFrames int64) error {
	if p.Status != StatusContinue {
		return ErrCannotBeUpdated
	}

	// Update current processed frames
	p.CurrentFrames = currentFrames

	// Calculate percentage using integer multiplication before division for precision.
	// min() is used to cap the value at 100 to handle FFmpeg metadata variances.
	pct := min(int((p.CurrentFrames*100)/p.TotalFrames), 100)

	p.Percentage = pct

	return nil
}

// End marks the progress as successfully completed.
func (p *Progress) End() error {
	if p.Status != StatusContinue {
		return ErrCannotBeMarkedAsEnd
	}

	p.Status = StatusEnd

	return nil
}

// MarkAsError marks the progress as failed due to a processing error.
func (p *Progress) MarkAsError() error {
	if p.Status != StatusContinue {
		return ErrCannotBeMarkedAsError
	}

	p.Status = StatusError

	return nil
}
