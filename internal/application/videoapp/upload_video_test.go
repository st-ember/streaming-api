package videoapp

import (
	"errors"
	"strings"
	"testing"

	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	storageMocks "github.com/st-ember/streaming-api/internal/application/ports/storage/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUploadVideo_SuccessCase(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// AssetStorer expectations
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(nil).
		Once()

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)

	mockUow.EXPECT().Rollback(mock.Anything).Return(nil) // will not run but expected due to defer func
	mockUow.EXPECT().Commit(mock.Anything).Return(nil).Once()

	// Repo expectations
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestUploadVideo_AssetStorerSaveFail(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// Expect AssetStorer Save to return error
	expectedErr := errors.New("path not found")
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(expectedErr).
		Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, resp)
}

func TestUploadVideo_UOWFactoryReturnsError(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// AssetStorer expectations
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(nil).
		Once()
	mockAsssetStorer.EXPECT().
		DeleteAll(mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	// Unit of Work Factory expectations
	expectedErr := errors.New("failed to connect to database")
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, expectedErr).
		Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, resp)
}

func TestUploadVideo_VideoRepoSaveFail(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// AssetStorer expectations
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(nil).
		Once()
	mockAsssetStorer.EXPECT().
		DeleteAll(mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)

	mockUow.EXPECT().Rollback(mock.Anything).Return(nil)

	// Repo expectations
	expectedErr := errors.New("duplicate key")
	mockVideoRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*video.Video")).
		Return(expectedErr).Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, resp)
}

func TestUploadVideo_JobRepoSaveFail(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// AssetStorer expectations
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(nil).
		Once()
	mockAsssetStorer.EXPECT().
		DeleteAll(mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)

	mockUow.EXPECT().Rollback(mock.Anything).Return(nil)

	// Repo expectations
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()

	expectedErr := errors.New("duplicate key")
	mockJobRepo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*job.Job")).
		Return(expectedErr).Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, resp)
}

func TestUploadVideo_UOWCommitFail(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockAsssetStorer := storageMocks.NewMockAssetStorer(t)
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockJobRepo := repoMocks.NewMockJobRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	// AssetStorer expectations
	mockAsssetStorer.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*strings.Reader")).
		Return(nil).
		Once()
	mockAsssetStorer.EXPECT().
		DeleteAll(mock.Anything, mock.AnythingOfType("string")).
		Return(nil)

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().JobRepo().Return(mockJobRepo)

	mockUow.EXPECT().Rollback(mock.Anything).Return(nil) // will not run but expected due to defer func

	expectedErr := errors.New("transaction already closed")
	mockUow.EXPECT().Commit(mock.Anything).Return(expectedErr).Once()

	// Repo expectations
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()
	mockJobRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*job.Job")).Return(nil).Once()

	// Mock input
	input := UploadVideoInput{
		Title:        "My Test Video",
		Description:  "A video for testing.",
		FileName:     "test.mp4",
		VideoContent: strings.NewReader("fake video data"),
	}
	// Create usecase
	usecase := NewUploadVideoUsecase(mockAsssetStorer, mockUowFactory)

	// Execute usecase
	resp, err := usecase.Execute(t.Context(), input)

	// --- Assert ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, resp)
}
