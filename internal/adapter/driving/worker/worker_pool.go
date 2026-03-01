package worker

import (
	"context"
	"sync"
	"time"

	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/storage"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type WorkerPool struct {
	startUC     jobapp.StartTranscodeJobUsecase
	completeUC  jobapp.CompleteTranscodeJobUsecase
	failUC      jobapp.FailTranscodeJobUsecase
	storer      storage.AssetStorer
	logger      log.Logger
	transcoder  transcode.Transcoder
	jobCh       chan *job.Job
	scheduler   *JobScheduler
	workerLimit int
	wg          sync.WaitGroup
}

func NewWorkerPool(
	findNextUC jobapp.FindNextPendingTranscodeJobUsecase,
	startUC jobapp.StartTranscodeJobUsecase,
	completeUC jobapp.CompleteTranscodeJobUsecase,
	failUC jobapp.FailTranscodeJobUsecase,
	storer storage.AssetStorer,
	logger log.Logger,
	transcoder transcode.Transcoder,
	pollInterval time.Duration,
	workerLimit int,
) *WorkerPool {
	jobCh := make(chan *job.Job, workerLimit)

	scheduler := NewJobScheduler(findNextUC, logger, jobCh, pollInterval, workerLimit)

	return &WorkerPool{
		startUC,
		completeUC,
		failUC,
		storer,
		logger,
		transcoder,
		jobCh,
		scheduler,
		workerLimit,
		sync.WaitGroup{},
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.scheduler.Run(ctx)
		close(p.jobCh)
	}()

	for range p.workerLimit {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			worker := NewTranscodeWorker(
				p.startUC, p.completeUC, p.failUC,
				p.storer, p.logger, p.transcoder, p.jobCh,
			)
			worker.Start()
		}()
	}
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
