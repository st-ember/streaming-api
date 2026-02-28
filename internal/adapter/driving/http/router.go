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
	logger log.Logger,
) *Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	videoH := handler.NewVideoHandler(uploadVideoUC, logger)

	// video
	videoRouter := api.PathPrefix("/video").Subrouter()
	videoRouter.HandleFunc("/", videoH.Upload).Methods(POST)

	return &Router{MuxRt: r}
}
