package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/st-ember/streaming-api/internal/adapter/driven/config"
	exec "github.com/st-ember/streaming-api/internal/adapter/driven/exec/os"
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
	// Setup Signal-aware Context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Config (use environment variables)
	cfg := config.Load()

	// Driven adapter (Repo)
	db, err := postgres.NewDB(cfg.ConnStr)
	if err != nil {
		log.Fatalf("start db connection: %v", err)
	}
	defer db.Conn.Close()

	uowFactory := postgres.NewPostgresUnitOfWorkFactory(db.Conn)

	// Driven adapter (Logger)
	logger := stdlib.NewStdLogger()

	// Driven adapter (Storer)
	storer, err := local.NewLocalAssetStorer(cfg.StoragePath)
	if err != nil {
		log.Fatalf("start storer: %v", err)
	}

	// Driven adapter (Exec Commander)
	execCommander := exec.NewOsCommander()
	transcoder := ffmpeg.NewFFMPEGTranscoder(cfg.StoragePath, execCommander)

	// Job Usecases
	completeTranscodeUC := jobapp.NewCompleteTranscodeJobUsecase(uowFactory)
	failTranscodeUC := jobapp.NewFailTranscodeJobUsecase(uowFactory)
	findNextUC := jobapp.NewFindNextPendingTranscodeJobUsecase(uowFactory)
	startTranscodeUC := jobapp.NewStartTranscodeJobUsecase(uowFactory)

	// Video Usecase
	uploadVideoUC := videoapp.NewUploadVideoUsecase(storer, uowFactory, logger)

	// Driving adapter (Worker)
	workerPool := worker.NewWorkerPool(
		findNextUC, startTranscodeUC, completeTranscodeUC, failTranscodeUC,
		storer, logger, transcoder, cfg.PollInterval, cfg.WorkerLimit,
	)
	workerPool.Start(ctx)

	// Driving adapter (HTTP)
	router := adpHttp.NewRouter(uploadVideoUC, logger)

	// Server config
	srv := &http.Server{
		Handler:      router.MuxRt,
		Addr:         ":" + cfg.ServerAdd,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start HTTP Server in background
	go func() {
		logger.Infof("Now listening on :%s\n", cfg.ServerAdd)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			stop() // initiate graceful shutdown
			logger.Errorf("listen: %s\n", err)
		}
	}()

	// Wait for signal to shut down
	<-ctx.Done()
	logger.Infof("shutting down gracefully...")

	// Shutdown server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	workerDone := make(chan struct{})
	go func() {
		workerPool.Wait()
		close(workerDone)
	}()

	select {
	case <-workerDone:
		logger.Infof("workers exited cleanly")
	case <-time.After(cfg.WorkerWaitTime):
		logger.Warnf("timed out waiting for workers; forcing exit")
	}

	logger.Infof("exiting")
}
