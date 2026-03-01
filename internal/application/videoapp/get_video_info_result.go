package videoapp

import "github.com/st-ember/streaming-api/internal/domain/video"

type GetVideoInfoResult struct {
	Video        *video.Video
	ManifestPath string
	ErrorMsg     string
}
