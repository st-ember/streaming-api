package http

import (
	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/adapter/driving/http/handler"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

type Router struct {
	MuxRt *mux.Router
}

var (
	GET    = "GET"
	POST   = "POST"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

func NewRouter(
	uploadVideoUC videoapp.UploadVideoUsecase,
	getInfoUC videoapp.GetVideoInfoUsecase,
	updateVideoUC videoapp.UpdateVideoUsecase,
	logger log.Logger,
) *Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	videoH := handler.NewVideoHandler(uploadVideoUC, getInfoUC, updateVideoUC, logger)

	// video
	videoRouter := api.PathPrefix("/video").Subrouter()
	videoRouter.HandleFunc("/", videoH.Upload).Methods(POST)
	videoRouter.HandleFunc("/{id}", videoH.Get).Methods(GET)
	videoRouter.HandleFunc("/{id}", videoH.Update).Methods(PATCH)

	return &Router{MuxRt: r}
}
