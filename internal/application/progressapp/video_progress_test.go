package progressapp_test

import (
	"errors"
	"testing"

	streamerMocks "github.com/st-ember/streaming-api/internal/application/ports/progressstream/mocks"
	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/application/progressapp"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/progress"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVideoProgress(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		// Set up mocks
		mockJobRepo := repoMocks.NewMockJobRepo(t)
		mockUow := repoMocks.NewMockUnitOfWork(t)
		mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)
		mockStreamer := streamerMocks.NewMockProgressStreamer(t)

		// Unit of Work Factory expectation
		mockUowFactory.EXPECT().
			NewUnitOfWork(mock.Anything).
			Return(mockUow, nil).
			Once()

		// Unit of Work expectation
		mockUow.EXPECT().JobRepo().Return(mockJobRepo)
		mockUow.EXPECT().Close(mock.Anything).Return(nil)

		// Repo expectation
		videoID := "test_video"
		jobID := "test_job"
		j := &job.Job{
			VideoID: videoID,
			ID:      jobID,
		}

		mockJobRepo.
			EXPECT().
			FindByVideoID(mock.Anything, videoID).
			Return(j, nil)

		// Progress streamer expectation
		expectedCh := make(<-chan *progress.Progress)
		mockStreamer.EXPECT().Read(mock.Anything, j.ID).Return(expectedCh, nil)

		// Create usecase
		usecase := progressapp.NewVideoProgressUsecase(mockStreamer, mockUowFactory)

		// Execute
		resultCh, err := usecase.Execute(t.Context(), videoID)

		// Assert
		require.NoError(t, err)
		require.Equal(t, expectedCh, resultCh)
	})

	t.Run("fails to initialize unit of work", func(t *testing.T) {
		// Set up mocks
		mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)
		mockStreamer := streamerMocks.NewMockProgressStreamer(t)

		// Unit of Work Factory expectation
		expectedErr := errors.New("db connection error")
		mockUowFactory.EXPECT().
			NewUnitOfWork(mock.Anything).
			Return(nil, expectedErr).
			Once()

		// Create usecase
		usecase := progressapp.NewVideoProgressUsecase(mockStreamer, mockUowFactory)

		// Execute
		resultCh, err := usecase.Execute(t.Context(), "test_video")

		// Assert
		require.ErrorContains(t, err, "initialize unit of work")
		require.Nil(t, resultCh)
	})

	t.Run("fails to find job", func(t *testing.T) {
		// Set up mocks
		mockJobRepo := repoMocks.NewMockJobRepo(t)
		mockUow := repoMocks.NewMockUnitOfWork(t)
		mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)
		mockStreamer := streamerMocks.NewMockProgressStreamer(t)

		// Unit of Work Factory expectation
		mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()

		// Unit of Work expectation
		mockUow.EXPECT().JobRepo().Return(mockJobRepo)
		mockUow.EXPECT().Close(mock.Anything).Return(nil)

		// Repo expectation
		videoID := "test_video"
		expectedErr := errors.New("not found")
		mockJobRepo.EXPECT().FindByVideoID(mock.Anything, videoID).Return(nil, expectedErr).Once()

		// Create usecase
		usecase := progressapp.NewVideoProgressUsecase(mockStreamer, mockUowFactory)

		// Execute
		resultCh, err := usecase.Execute(t.Context(), videoID)

		// Assert
		require.ErrorContains(t, err, "find job with id")
		require.Nil(t, resultCh)
	})

	t.Run("fails to read from streamer", func(t *testing.T) {
		// Set up mocks
		mockJobRepo := repoMocks.NewMockJobRepo(t)
		mockUow := repoMocks.NewMockUnitOfWork(t)
		mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)
		mockStreamer := streamerMocks.NewMockProgressStreamer(t)

		// Unit of Work Factory expectation
		mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()

		// Unit of Work expectation
		mockUow.EXPECT().JobRepo().Return(mockJobRepo)
		mockUow.EXPECT().Close(mock.Anything).Return(nil)

		// Repo expectation
		videoID := "test_video"
		jobID := "test_job"
		j := &job.Job{ID: jobID}

		mockJobRepo.EXPECT().FindByVideoID(mock.Anything, videoID).Return(j, nil).Once()

		// Progress streamer expectation
		expectedErr := errors.New("redis error")
		mockStreamer.EXPECT().Read(mock.Anything, jobID).Return(nil, expectedErr).Once()

		// Create usecase
		usecase := progressapp.NewVideoProgressUsecase(mockStreamer, mockUowFactory)

		// Execute
		resultCh, err := usecase.Execute(t.Context(), videoID)

		// Assert
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, resultCh)
	})
}
