package videoapp_test

import (
	"errors"
	"testing"

	repoMocks "github.com/st-ember/streaming-api/internal/application/ports/repo/mocks"
	"github.com/st-ember/streaming-api/internal/application/videoapp"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateVideo_SuccessPartial(t *testing.T) {
	t.Parallel()

	// Set up mocks
	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	resourceID := "res-123"
	oldTitle := "Old Title"
	oldDesc := "Old Desc"

	// Existing video state
	testVideo, _ := video.NewVideo(videoID, oldTitle, oldDesc, "test.mp4", resourceID)

	// Unit of Work Factory expectations
	mockUowFactory.EXPECT().
		NewUnitOfWork(mock.Anything).
		Return(mockUow, nil).
		Once()

	// Unit of Work expectations
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()
	mockUow.EXPECT().Commit(mock.Anything).Return(nil).Once()

	// Repo expectations: Fetch then Save
	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()
	mockVideoRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*video.Video")).Return(nil).Once()

	// Partial Update: Only new title
	newTitle := "New Awesome Title"
	input := videoapp.UpdateVideoInput{
		ID:    videoID,
		Title: &newTitle,
	}

	usecase := videoapp.NewUpdateVideoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, newTitle, result.Title)
	require.Equal(t, oldDesc, result.Description) // Description remains unchanged
}

func TestUpdateVideo_VideoNotFound(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "non-existent"
	expectedErr := errors.New("not found")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(nil, expectedErr).Once()

	input := videoapp.UpdateVideoInput{ID: videoID}
	usecase := videoapp.NewUpdateVideoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), input)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), expectedErr.Error())
}

func TestUpdateVideo_DomainValidationError(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	videoID := "video-123"
	testVideo, _ := video.NewVideo(videoID, "Title", "Desc", "test.mp4", "res-123")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().FindByID(mock.Anything, videoID).Return(testVideo, nil).Once()

	// Trying to update to an empty title (domain validation should fail)
	invalidTitle := ""
	input := videoapp.UpdateVideoInput{
		ID:    videoID,
		Title: &invalidTitle,
	}

	usecase := videoapp.NewUpdateVideoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), input)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, video.ErrTitleEmpty) // Should reflect domain error
}
