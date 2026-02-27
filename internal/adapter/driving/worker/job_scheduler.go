package worker

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type JobScheduler struct {
	findNextUC jobapp.FindNextPendingTranscodeJobUsecase
	logger     log.Logger
	jobCh      chan *job.Job
}

func NewJobScheduler(
	findNextUC jobapp.FindNextPendingTranscodeJobUsecase,
	logger log.Logger,
	jobCh chan *job.Job,
) *JobScheduler {
	return &JobScheduler{
		findNextUC,
		logger,
		jobCh,
	}
}

var workerLimit = 5

func (s *JobScheduler) Run(ctx context.Context) {
	s.logger.Infof("starting worker pool")

	waitSec := 10
	ticker := time.NewTicker(time.Duration(waitSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Infof("shutting down worker pool")
			return
		case <-ticker.C:
			job, err := s.findNextUC.Execute(ctx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				s.logger.Errorf("find next executable job: %v", err)
				time.Sleep(5 * time.Second) // backoff
			}
			if job == nil {
				continue
			}

			select {
			// Send job
			case s.jobCh <- job:
				s.logger.Infof("job %s is added to queue", job.ID)
			default: // Default case to make the scheduler more reactive for later adjustments
				s.logger.Infof("job queue full now, will try again in %v seconds", waitSec)
			}
		}
	}
}
