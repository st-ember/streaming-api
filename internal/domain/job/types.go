package job

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

type JobType string

// Potential type for thumbnail generation
const TypeTranscode JobType = "transcode"

func (jt JobType) IsValid() bool {
	switch jt {
	case TypeTranscode:
		return true
	default:
		return false
	}
}
