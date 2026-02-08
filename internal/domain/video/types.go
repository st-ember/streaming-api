package video

type VideoStatus string

const (
	StatusPending    VideoStatus = "pending"
	StatusProcessing VideoStatus = "processing"
	StatusPublished  VideoStatus = "published"
	StatusFailed     VideoStatus = "failed"
	StatusArchived   VideoStatus = "archived"
)
