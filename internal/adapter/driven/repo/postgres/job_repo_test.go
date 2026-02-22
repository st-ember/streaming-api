package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/stretchr/testify/require"
)

func TestPostgresJobRepo_Save_Insert(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresJobRepo(tx)
	newJob, err := job.NewJob("job-id-1", "video-id-1", job.TypeTranscode)
	require.NoError(t, err)

	// ACT
	err = repo.Save(t.Context(), newJob)

	// require
	require.NoError(t, err)

	// Verify by reading the data back directly from the database transaction.
	var status string
	err = tx.QueryRow("SELECT status FROM jobs WHERE id = $1", "job-id-1").Scan(&status)
	require.NoError(t, err)
	require.Equal(t, string(job.StatusPending), status)
}

func TestPostgresJobRepo_Save_Update(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresJobRepo(tx)

	// First, insert a job.
	originalJob, err := job.NewJob("job-id-1", "video-id-1", job.TypeTranscode)
	require.NoError(t, err)
	err = repo.Save(t.Context(), originalJob)
	require.NoError(t, err)

	// Now, modify the job object.
	originalJob.Status = job.StatusRunning
	originalJob.ErrorMsg = "an error occurred"
	time.Sleep(1 * time.Nanosecond) // ensure updated_at changes
	originalJob.UpdatedAt = time.Now()

	// ACT
	// Call Save again on the same object. It should perform an UPDATE.
	err = repo.Save(t.Context(), originalJob)

	// require
	require.NoError(t, err)

	// Verify the update by reading the data back.
	var updatedStatus, updatedErrorMsg string
	err = tx.QueryRow("SELECT status, error_msg FROM jobs WHERE id = $1", "job-id-1").Scan(&updatedStatus, &updatedErrorMsg)
	require.NoError(t, err)
	require.Equal(t, string(job.StatusRunning), updatedStatus)
	require.Equal(t, "an error occurred", updatedErrorMsg)
}

func TestPostgresJobRepo_FindNextPendingTranscodeJob_Success(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresJobRepo(tx)

	// Insert some test jobs directly into the database.
	// This job should NOT be picked.
	_, err := tx.Exec(`INSERT INTO jobs (id, video_id, type, status, created_at, updated_at) 
		VALUES ('job-1', 'vid-1', 'transcode', 'running', $1, $1)`, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	// This is the oldest pending transcode job, it SHOULD be picked.
	oldestPendingTime := time.Now().Add(-30 * time.Minute)
	_, err = tx.Exec(`INSERT INTO jobs (id, video_id, type, status, created_at, updated_at) 
		VALUES ('job-2-oldest', 'vid-2', 'transcode', 'pending', $1, $1)`, oldestPendingTime)
	require.NoError(t, err)

	// This job is pending, but newer, so it should NOT be picked.
	_, err = tx.Exec(`INSERT INTO jobs (id, video_id, type, status, created_at, updated_at) 
		VALUES ('job-3-newer', 'vid-3', 'transcode', 'pending', $1, $1)`, time.Now())
	require.NoError(t, err)

	// This job is pending, but not a transcode job, so it should NOT be picked.
	_, err = tx.Exec(`INSERT INTO jobs (id, video_id, type, status, created_at, updated_at) 
		VALUES ('job-4-thumbnail', 'vid-4', 'thumbnail', 'pending', $1, $1)`, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	// ACT
	foundJob, err := repo.FindNextPendingTranscodeJob(t.Context())

	// require
	require.NoError(t, err)
	require.NotNil(t, foundJob)
	require.Equal(t, "job-2-oldest", foundJob.ID) // Verify we found the correct job.
}

func TestPostgresJobRepo_FindNextPendingTranscodeJob_NotFound(t *testing.T) {
	t.Parallel()
	tx := beginTx(t)

	// ARRANGE
	repo := NewPostgresJobRepo(tx)
	// Insert jobs, but none that are pending transcodes.
	_, err := tx.Exec(`INSERT INTO jobs (id, video_id, type, status, created_at, updated_at) 
		VALUES ('job-1', 'vid-1', 'transcode', 'running', $1, $1)`, time.Now())
	require.NoError(t, err)

	// ACT
	foundJob, err := repo.FindNextPendingTranscodeJob(t.Context())

	// require
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, foundJob)
}
