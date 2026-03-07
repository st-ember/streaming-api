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
	videoUC videoapp.VideoUsecase,
	logger log.Logger,
) *Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	videoH := handler.NewVideoHandler(videoUC, logger)

	// video
	videoRouter := api.PathPrefix("/video").Subrouter()
	videoRouter.HandleFunc("/", videoH.Upload).Methods(POST)
	videoRouter.HandleFunc("/{id}", videoH.Get).Methods(GET)
	videoRouter.HandleFunc("/{id}", videoH.Update).Methods(PATCH)
	videoRouter.HandleFunc("/{id}", videoH.Archive).Methods(DELETE)
	videoRouter.HandleFunc("/", videoH.List).Methods(GET).Queries("page", "{page:[0-9]+}")

	return &Router{MuxRt: r}
}
