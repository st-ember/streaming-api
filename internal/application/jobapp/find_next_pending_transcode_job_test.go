package jobapp

import (
	"errors"
	"testing"

	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindNextPendingTranscodeJob_SuccessCase(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockJobRepo := repoMocks.NewMockJobRepo(t)

	// Create a job entity to return
	expectedJob, err := job.NewJob(
		"mock_job_id",
		"mock_video_id",
		job.TypeTranscode,
	)
	require.NoError(t, err)

	// Job repo expectations
	mockJobRepo.EXPECT().FindNextPendingTranscodeJob(mock.Anything).Return(expectedJob, nil).Once()

	// Create usecase
	usecase := NewFindNextPendingTranscodeJobUsecase(mockJobRepo)

	// Execute usecase
	foundJob, err := usecase.Execute(t.Context())

	// ---Assertions---
	require.NoError(t, err)
	require.NotNil(t, foundJob)
	require.Equal(t, expectedJob.ID, foundJob.ID)
	require.Equal(t, expectedJob.VideoID, foundJob.VideoID)
	require.Equal(t, expectedJob.Type, foundJob.Type)
}

func TestFindNextPendingTranscodeJob_JobRepoReturnsError(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockJobRepo := repoMocks.NewMockJobRepo(t)

	// Job repo expectations
	expectedErr := errors.New("connection failed")
	mockJobRepo.EXPECT().FindNextPendingTranscodeJob(mock.Anything).Return(nil, expectedErr).Once()

	// Create usecase
	usecase := NewFindNextPendingTranscodeJobUsecase(mockJobRepo)

	// Execute usecase
	foundJob, err := usecase.Execute(t.Context())

	// ---Assertions---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, foundJob)
}

func TestFindNextPendingTranscodeJob_NoJobFound(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockJobRepo := repoMocks.NewMockJobRepo(t)

	// Job repo expectations
	mockJobRepo.EXPECT().FindNextPendingTranscodeJob(mock.Anything).Return(nil, nil).Once()

	// Create usecase
	usecase := NewFindNextPendingTranscodeJobUsecase(mockJobRepo)

	// Execute usecase
	foundJob, err := usecase.Execute(t.Context())

	// ---Assertions---
	require.NoError(t, err)
	require.Nil(t, foundJob)
}
