package main

import (
	"context"
	"log"

	"github.com/st-ember/streaming-api/internal/adapter/driven/exec/os"
	"github.com/st-ember/streaming-api/internal/adapter/driven/log/stdlib"
	"github.com/st-ember/streaming-api/internal/adapter/driven/repo/postgres"
	"github.com/st-ember/streaming-api/internal/adapter/driven/storage/local"
	"github.com/st-ember/streaming-api/internal/adapter/driven/transcode/ffmpeg"
	"github.com/st-ember/streaming-api/internal/adapter/driving/worker"
	"github.com/st-ember/streaming-api/internal/application/jobapp"
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

	// Driving adapters
	workerPool := worker.NewWorkerPool(
		findNextUC, startTranscodeUC, completeTranscodeUC,
		failTranscodeUC, storer, logger, transcoder, 5,
	)

	workerPool.Start(ctx)
}
