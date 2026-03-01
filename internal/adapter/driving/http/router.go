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
	PUT    = "PUT"
	DELETE = "DELETE"
)

func NewRouter(
	uploadVideoUC videoapp.UploadVideoUsecase,
	getInfoUC videoapp.GetVideoInfoUsecase,
	logger log.Logger,
) *Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	videoH := handler.NewVideoHandler(uploadVideoUC, getInfoUC, logger)

	// video
	videoRouter := api.PathPrefix("/video").Subrouter()
	videoRouter.HandleFunc("/", videoH.Upload).Methods(POST)
	videoRouter.HandleFunc("/{id}", videoH.Get).Methods(GET)

	return &Router{MuxRt: r}
}
