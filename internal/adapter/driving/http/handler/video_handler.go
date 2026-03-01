package handler

import (
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

type VideoHandler struct {
	uploadVideoUC videoapp.UploadVideoUsecase
	getVideoUC    videoapp.GetVideoInfoUsecase
	logger        log.Logger
}

func NewVideoHandler(
	uploadVideoUC videoapp.UploadVideoUsecase,
	getVideoUC videoapp.GetVideoInfoUsecase,
	logger log.Logger,
) *VideoHandler {
	return &VideoHandler{
		uploadVideoUC,
		getVideoUC,
		logger,
	}
}
