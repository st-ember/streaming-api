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

func TestListVideos_Success(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	page := 1
	expectedVideos := []*video.Video{
		{ID: "video-1", Title: "First Video"},
		{ID: "video-2", Title: "Second Video"},
	}

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().List(mock.Anything, page).Return(expectedVideos, nil).Once()

	usecase := videoapp.NewListVideoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), page)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "video-1", result[0].ID)
}

func TestListVideos_RepoError(t *testing.T) {
	t.Parallel()

	mockVideoRepo := repoMocks.NewMockVideoRepo(t)
	mockUow := repoMocks.NewMockUnitOfWork(t)
	mockUowFactory := repoMocks.NewMockUnitOfWorkFactory(t)

	page := 1
	expectedErr := errors.New("db connection lost")

	mockUowFactory.EXPECT().NewUnitOfWork(mock.Anything).Return(mockUow, nil).Once()
	mockUow.EXPECT().VideoRepo().Return(mockVideoRepo)
	mockUow.EXPECT().Rollback(mock.Anything).Return(nil).Once()

	mockVideoRepo.EXPECT().List(mock.Anything, page).Return(nil, expectedErr).Once()

	usecase := videoapp.NewListVideoUsecase(mockUowFactory)
	result, err := usecase.Execute(t.Context(), page)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, expectedErr)
}
