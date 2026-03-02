package handler

import (
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

type VideoHandler struct {
	uploadVideoUC videoapp.UploadVideoUsecase
	getVideoUC    videoapp.GetVideoInfoUsecase
	updateVideoUC videoapp.UpdateVideoUsecase
	logger        log.Logger
}

func NewVideoHandler(
	uploadVideoUC videoapp.UploadVideoUsecase,
	getVideoInfoUC videoapp.GetVideoInfoUsecase,
	updateVideuoUC videoapp.UpdateVideoUsecase,
	logger log.Logger,
) *VideoHandler {
	return &VideoHandler{
		uploadVideoUC,
		getVideoInfoUC,
		updateVideuoUC,
		logger,
	}
}
