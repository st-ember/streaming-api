package postgres

import (
	"database/sql"
	"testing"

	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/require"
)

func TestUnitOfWork_Commit(t *testing.T) {
	t.Parallel()
	// ARRANGE
	// Note: For UoW tests, we need the factory which holds the main DB pool,
	// not just a single transaction from beginTx.
	uowFactory := NewPostgresUnitOfWorkFactory(testDB)

	// Create a video entity to save.
	newVideo, err := video.NewVideo("uow-video-1", "Commit Test", "Desc", "commit.mp4", "uow-resource-1")
	require.NoError(t, err)

	// ACT
	// Start a new Unit of Work.
	uow, err := uowFactory.NewUnitOfWork(t.Context())
	require.NoError(t, err)

	// Get the transactional repo and save the entity.
	videoRepo := uow.VideoRepo()
	err = videoRepo.Save(t.Context(), newVideo)
	require.NoError(t, err)

	// Commit the transaction.
	err = uow.Commit(t.Context())
	require.NoError(t, err)

	// ASSERT
	// Verify that the data was actually committed to the database.
	// We do this by starting a NEW, separate query against the main connection pool.
	var title string
	err = testDB.QueryRow("SELECT title FROM videos WHERE id = $1", newVideo.ID).Scan(&title)
	require.NoError(t, err, "data should be present in the DB after commit")
	require.Equal(t, "Commit Test", title)
}

func TestUnitOfWork_Rollback(t *testing.T) {
	t.Parallel()
	// ARRANGE
	uowFactory := NewPostgresUnitOfWorkFactory(testDB)
	ctx := t.Context() // Use t.Context() for calls within the test logic.

	newVideo, err := video.NewVideo("uow-video-2", "Rollback Test", "Desc", "rollback.mp4", "uow-resource-2")
	require.NoError(t, err)

	// ACT
	// Start a new Unit of Work.
	uow, err := uowFactory.NewUnitOfWork(ctx)
	require.NoError(t, err)

	// Save an entity within the transaction.
	err = uow.VideoRepo().Save(ctx, newVideo)
	require.NoError(t, err)

	// Immediately roll back the transaction.
	err = uow.Rollback(ctx)
	require.NoError(t, err)

	// ASSERT
	// Verify that the data was NOT committed to the database.
	// This query against the main connection pool should find nothing.
	var title string
	err = testDB.QueryRow("SELECT title FROM videos WHERE id = $1", newVideo.ID).Scan(&title)
	require.ErrorIs(t, err, sql.ErrNoRows, "data should NOT be present in the DB after rollback")
}

func TestUnitOfWork_RepoIsTransactional(t *testing.T) {
	t.Parallel()
	// ARRANGE
	uowFactory := NewPostgresUnitOfWorkFactory(testDB)
	ctx := t.Context()

	newVideo, err := video.NewVideo("uow-video-3", "TX Test", "Desc", "tx.mp4", "uow-resource-3")
	require.NoError(t, err)
	newJob, err := job.NewJob("uow-job-3", newVideo.ID, job.TypeTranscode)
	require.NoError(t, err)

	// ACT & ASSERT
	// Start a new Unit of Work.
	uow, err := uowFactory.NewUnitOfWork(ctx)
	require.NoError(t, err)

	// Get both repositories from the SAME Unit of Work.
	videoRepo := uow.VideoRepo()
	jobRepo := uow.JobRepo()

	// Save both entities using their respective repositories.
	err = videoRepo.Save(ctx, newVideo)
	require.NoError(t, err)
	err = jobRepo.Save(ctx, newJob)
	require.NoError(t, err)

	// At this point, the data should exist WITHIN the transaction, but not outside it.
	// We can't easily test this without a second connection, so we'll proceed to rollback.

	// Now, roll back the entire transaction.
	err = uow.Rollback(ctx)
	require.NoError(t, err)

	// Verify that BOTH saves were rolled back, proving they were in the same transaction.
	var videoID string
	err = testDB.QueryRow("SELECT id FROM videos WHERE id = $1", newVideo.ID).Scan(&videoID)
	require.ErrorIs(t, err, sql.ErrNoRows, "video should have been rolled back")

	var jobID string
	err = testDB.QueryRow("SELECT id FROM jobs WHERE id = $1", newJob.ID).Scan(&jobID)
	require.ErrorIs(t, err, sql.ErrNoRows, "job should have been rolled back")
}
