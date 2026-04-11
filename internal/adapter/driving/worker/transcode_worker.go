package worker

import (
	"context"
	"os"
	"path/filepath"

	"github.com/st-ember/streaming-api/internal/application/jobapp"
	"github.com/st-ember/streaming-api/internal/application/ports/log"
	"github.com/st-ember/streaming-api/internal/application/ports/storage"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
	"github.com/st-ember/streaming-api/internal/domain/job"
)

type TranscodeWorker struct {
	startUC    jobapp.StartTranscodeJobUsecase
	completeUC jobapp.CompleteTranscodeJobUsecase
	failUC     jobapp.FailTranscodeJobUsecase
	storer     storage.AssetStorer
	logger     log.Logger
	transcoder transcode.Transcoder
	jobCh      chan *job.Job
}

func NewTranscodeWorker(
	startUC jobapp.StartTranscodeJobUsecase,
	completeUC jobapp.CompleteTranscodeJobUsecase,
	failUC jobapp.FailTranscodeJobUsecase,
	storer storage.AssetStorer,
	logger log.Logger,
	transcoder transcode.Transcoder,
	jobCh chan *job.Job,
) *TranscodeWorker {
	return &TranscodeWorker{
		startUC,
		completeUC,
		failUC,
		storer,
		logger,
		transcoder,
		jobCh,
	}
}

func (w *TranscodeWorker) Start(ctx context.Context) {
	for job := range w.jobCh {
		func() {
			resp, err := w.startUC.Execute(ctx, job)
			if err != nil {
				w.logger.Errorf(ctx, "start job %s: %v", job.ID, err)
				return
			}

			out, err := w.transcoder.Transcode(ctx, resp.ResourceID, resp.SourceFilename, job.ID)
			if err != nil {
				// Execute fail transcode job usecase
				w.failUC.Execute(ctx, job, err.Error())
				w.logger.Errorf(ctx, "transcode job %s: %v", job.ID, err)
				return
			}

			tempOutputDir := filepath.Dir(out.ManifestPath)
			// Clean up temporary directory on error or when job finishes
			defer func() {
				if err := os.RemoveAll(tempOutputDir); err != nil {
					w.logger.Errorf(ctx, "job %s: clean up temporary directory %s: %v", job.ID, tempOutputDir, err)
				}
			}()

			// Move temp files from manifest path into permanent storage
			for _, relativeFilePath := range out.OutputFiles {
				fullTempPath := filepath.Join(tempOutputDir, relativeFilePath)
				tempFile, err := os.Open(fullTempPath)
				if err != nil {
					w.logger.Errorf(ctx, "job %s: open temporary file %s for saving: %v", job.ID, fullTempPath, err)
					w.failUC.Execute(ctx, job, "failed to read transcoded output")
					return
				}

				err = w.storer.Save(ctx, resp.ResourceID, relativeFilePath, tempFile)
				// Close immediately
				tempFile.Close()
				if err != nil {
					w.logger.Errorf(ctx, "job %s: save transcoded file %s to storage: %v", job.ID, relativeFilePath, err)
					w.failUC.Execute(ctx, job, "failed to save transcoded output")
					return
				}

				// Log successful move
				w.logger.Infof(ctx, "deleted and moved temp files to permanent storage for video %s", resp.ResourceID)
			}

			if err := w.completeUC.Execute(ctx, job, filepath.Base(out.ManifestPath), out.Duration); err != nil {
				w.logger.Errorf(ctx, "complete job %s: %v", job.ID, err)
			}

			// Log successful job completion
			w.logger.Infof(ctx, "completed job %s", job.ID)
		}()
	}

	w.logger.Infof(ctx, "worker finished draining queue and is shutting down")
}
