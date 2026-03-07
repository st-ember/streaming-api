package handler

import (
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

type VideoHandler struct {
	videoUC videoapp.VideoUsecase
	logger  log.Logger
}

func NewVideoHandler(
	videoUC videoapp.VideoUsecase,
	logger log.Logger,
) *VideoHandler {
	return &VideoHandler{
		videoUC,
		logger,
	}
}
