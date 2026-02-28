package videoapp

import (
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
)

type UploadVideoResult struct {
	Video *video.Video
	Job   *job.Job
}
