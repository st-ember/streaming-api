package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/require"
)

func TestPostgresVideoRepo_Save_Insert(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresVideoRepo(tx)
	newVideo, err := video.NewVideo("video-id-1", "Test Title", "Test Desc", "test.mp4", "resource-1")
	require.NoError(t, err)

	// ACT
	err = repo.Save(t.Context(), newVideo)

	// ASSERT
	require.NoError(t, err)

	// Verify by reading the data back directly from the database transaction.
	var title string
	err = tx.QueryRow("SELECT title FROM videos WHERE id = $1", "video-id-1").Scan(&title)
	require.NoError(t, err)
	require.Equal(t, "Test Title", title)
}

func TestPostgresVideoRepo_Save_Update(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresVideoRepo(tx)

	// First, insert a video.
	originalVideo, err := video.NewVideo("video-id-1", "Original Title", "Original Desc", "test.mp4", "resource-1")
	require.NoError(t, err)
	err = repo.Save(t.Context(), originalVideo)
	require.NoError(t, err)

	// Now, modify the video object.
	originalVideo.Title = "Updated Title"
	originalVideo.Status = video.StatusPublished
	time.Sleep(1 * time.Nanosecond) // ensure updated_at changes
	originalVideo.UpdatedAt = time.Now()

	// ACT
	// Call Save again on the same object. It should perform an UPDATE.
	err = repo.Save(t.Context(), originalVideo)

	// ASSERT
	require.NoError(t, err)

	// Verify the update by reading the data back.
	var updatedTitle string
	var updatedStatus string
	err = tx.QueryRow("SELECT title, status FROM videos WHERE id = $1", "video-id-1").Scan(&updatedTitle, &updatedStatus)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", updatedTitle)
	require.Equal(t, string(video.StatusPublished), updatedStatus)
}

func TestPostgresVideoRepo_FindByID_Success(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresVideoRepo(tx)
	// Insert a video to be found.
	videoToFind, err := video.NewVideo("video-id-2", "Find Me", "Desc", "find.mp4", "resource-2")
	require.NoError(t, err)
	err = repo.Save(t.Context(), videoToFind)
	require.NoError(t, err)

	// ACT
	foundVideo, err := repo.FindByID(t.Context(), "video-id-2")

	// ASSERT
	require.NoError(t, err)
	require.NotNil(t, foundVideo)
	require.Equal(t, videoToFind.ID, foundVideo.ID)
	require.Equal(t, videoToFind.Title, foundVideo.Title)
}

func TestPostgresVideoRepo_FindByID_NotFound(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresVideoRepo(tx)

	// ACT
	// Try to find a video that does not exist.
	foundVideo, err := repo.FindByID(t.Context(), "non-existent-id")

	// ASSERT
	require.ErrorIs(t, err, sql.ErrNoRows) // Verify the specific "not found" error is returned.
	require.Nil(t, foundVideo)
}
