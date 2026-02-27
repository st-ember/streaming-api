package worker

import (
	"context"
	"time"

	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/storage"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type WorkerPool struct {
	startUC    jobapp.StartTranscodeJobUsecase
	completeUC jobapp.CompleteTranscodeJobUsecase
	failUC     jobapp.FailTranscodeJobUsecase
	storer     storage.AssetStorer
	logger     log.Logger
	transcoder transcode.Transcoder
	jobCh      chan *job.Job
	scheduler  *JobScheduler
}

func NewWorkerPool(
	findNextUC jobapp.FindNextPendingTranscodeJobUsecase,
	startUC jobapp.StartTranscodeJobUsecase,
	completeUC jobapp.CompleteTranscodeJobUsecase,
	failUC jobapp.FailTranscodeJobUsecase,
	storer storage.AssetStorer,
	logger log.Logger,
	transcoder transcode.Transcoder,
	jobCh chan *job.Job,
) *WorkerPool {
	scheduler := NewJobScheduler(findNextUC, logger, jobCh, 10*time.Second)

	return &WorkerPool{
		startUC,
		completeUC,
		failUC,
		storer,
		logger,
		transcoder,
		jobCh,
		scheduler,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	go p.scheduler.Run(ctx)

	for range workerLimit {
		go func() {
			worker := NewTranscodeWorker(
				p.startUC, p.completeUC, p.failUC,
				p.storer, p.logger, p.transcoder, p.jobCh,
			)
			worker.Start(ctx)
		}()
	}
}
