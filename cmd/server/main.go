package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/st-ember/streaming-api/internal/adapter/driven/exec/os"
	"github.com/st-ember/streaming-api/internal/adapter/driven/log/stdlib"
	"github.com/st-ember/streaming-api/internal/adapter/driven/repo/postgres"
	"github.com/st-ember/streaming-api/internal/adapter/driven/storage/local"
	"github.com/st-ember/streaming-api/internal/adapter/driven/transcode/ffmpeg"
	adpHttp "github.com/st-ember/streaming-api/internal/adapter/driving/http"
	"github.com/st-ember/streaming-api/internal/adapter/driving/worker"
	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
)

func main() {
	// Context
	ctx := context.Background()

	// Driven adapters
	db, err := postgres.NewDB("")
	if err != nil {
		log.Fatalf("start db connection: %v", err)
	}

	uowFactory := postgres.NewPostgresUnitOfWorkFactory(db.Conn)

	storer, err := local.NewLocalAssetStorer("")
	if err != nil {
		log.Fatalf("start storer: %v", err)
	}

	logger := stdlib.NewStdLogger()
	execCommander := os.NewOsCommander()
	transcoder := ffmpeg.NewFFMPEGTranscoder("", execCommander)

	// Usecases
	completeTranscodeUC := jobapp.NewCompleteTranscodeJobUsecase(uowFactory)
	failTranscodeUC := jobapp.NewFailTranscodeJobUsecase(uowFactory)
	findNextUC := jobapp.NewFindNextPendingTranscodeJobUsecase(uowFactory)
	startTranscodeUC := jobapp.NewStartTranscodeJobUsecase(uowFactory)

	uploadVideoUC := videoapp.NewUploadVideoUsecase(storer, uowFactory, logger)

	// Driving adapters
	workerPool := worker.NewWorkerPool(
		findNextUC, startTranscodeUC, completeTranscodeUC,
		failTranscodeUC, storer, logger, transcoder, 5,
	)

	workerPool.Start(ctx)

	router := adpHttp.NewRouter(uploadVideoUC, logger)

	serverAdd := "8085"

	srv := &http.Server{
		Handler: router.MuxRt,
		Addr:    serverAdd,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Now listening on %s", serverAdd)

	log.Fatal(srv.ListenAndServe())
}
