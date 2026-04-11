package http

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/st-ember/streaming-api/internal/adapter/driving/http/handler"
	wshandler "github.com/st-ember/streaming-api/internal/adapter/driving/websocket/handler"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/progressapp"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

type Router struct {
	MuxRt   *mux.Router
	Handler http.Handler
}

var (
	GET    = "GET"
	POST   = "POST"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

func NewRouter(
	videoUC videoapp.VideoUsecase,
	videoProgressUC progressapp.VideoProgressUsecase,
	storagePath string,
	allowedCfg []string,
	logger log.Logger,
) *Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	// video
	videoRouter := api.PathPrefix("/video").Subrouter()
	videoH := handler.NewVideoHandler(videoUC, logger)
	videoRouter.HandleFunc("/", videoH.Upload).Methods(POST)
	videoRouter.HandleFunc("/{id}", videoH.Get).Methods(GET)
	videoRouter.HandleFunc("/{id}", videoH.Update).Methods(PATCH)
	videoRouter.HandleFunc("/{id}", videoH.Archive).Methods(DELETE)
	videoRouter.HandleFunc("/list/{page}", videoH.List).Methods(GET)

	// streaming
	streamingRouter := r.PathPrefix("/streaming").Subrouter()
	streamingHandler := handler.NewStreamingHandler(storagePath, logger)
	streamingRouter.HandleFunc("/{resourceID}/{filename}", streamingHandler.ServeFile).Methods(GET)

	// progress
	progressRouter := r.PathPrefix("progress").Subrouter()
	progressHandler := wshandler.NewProgressHandler(videoProgressUC, logger)
	progressRouter.HandleFunc("/video/{id}", progressHandler.VideoProgress).Methods(GET)
	// later thumbnail generation progress handler may be added

	// cors config
	allowedOrigins := handlers.AllowedOrigins(allowedCfg)

	// apply router to cors handler
	corsHandler := handlers.CORS(allowedOrigins)(r)

	return &Router{MuxRt: r, Handler: corsHandler}
}
