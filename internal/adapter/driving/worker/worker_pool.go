package worker

import (
	"context"

	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/storage"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type WorkerPool struct {
	startUC    *jobapp.StartTranscodeJobUsecase
	completeUC *jobapp.CompleteTranscodeJobUsecase
	failUC     *jobapp.FailTranscodeJobUsecase
	storer     storage.AssetStorer
	logger     log.Logger
	transcoder transcode.Transcoder
	jobCh      chan *job.Job
}

func NewWorkerPool(
	findNextUC *jobapp.FindNextPendingTranscodeJobUsecase,
	startUC *jobapp.StartTranscodeJobUsecase,
	completeUC *jobapp.CompleteTranscodeJobUsecase,
	failUC *jobapp.FailTranscodeJobUsecase,
	storer storage.AssetStorer,
	logger log.Logger,
	transcoder transcode.Transcoder,
	jobCh chan *job.Job,
) *WorkerPool {
	return &WorkerPool{
		startUC,
		completeUC,
		failUC,
		storer,
		logger,
		transcoder,
		jobCh,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
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
