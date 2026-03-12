package job_test

import (
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/domain/job"
	"github.com/stretchr/testify/require"
)

// jobTestHelper encapsulates common test setup and assertions.
type jobTestHelper struct {
	*require.Assertions
	mockID      string
	mockVideoID string
	mockJobType job.JobType
}

// setupJobTestHelper creates a new helper for a given test.
func setupJobTestHelper(t *testing.T) *jobTestHelper {
	return &jobTestHelper{
		Assertions:  require.New(t),
		mockID:      "mock_job_id",
		mockVideoID: "mock_video_id",
		mockJobType: job.TypeTranscode,
	}
}

func TestNewJob_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)

	h.NoError(err)
	h.NotNil(j)
	h.Equal(h.mockID, j.ID)
	h.Equal(h.mockVideoID, j.VideoID)
	h.Equal(h.mockJobType, j.Type)
	h.Equal(job.StatusPending, j.Status)
}

func TestNewJob_FailsOnEmptyID(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob("", h.mockVideoID, h.mockJobType)

	h.Nil(j)
	h.ErrorIs(err, job.ErrJobIDEmpty)
}

func TestNewJob_FailsOnEmptyVideoID(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, "", h.mockJobType)

	h.Nil(j)
	h.ErrorIs(err, job.ErrVideoIDEmpty)
}

func TestNewJob_FailsOnInvalidJobType(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, job.JobType("invalid_type"))

	h.Nil(j)
	h.ErrorIs(err, job.ErrJobTypeInvalid)
}

func TestStart_SuccessCaseFromPending(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(job.StatusPending, j.Status)

	err = j.Start()
	h.NoError(err)
	h.Equal(job.StatusRunning, j.Status)
}

func TestStart_SuccessCaseFromFailed(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	j.Status = job.StatusFailed // Manually set state for test

	err = j.Start()
	h.NoError(err)
	h.Equal(job.StatusRunning, j.Status)
}

func TestStart_FailsIfNotPendingOrFailed(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	j.Status = job.StatusRunning // Set to a non-startable state

	err = j.Start()
	h.ErrorIs(err, job.ErrCannotBeStarted)
}

func TestComplete_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	j.Status = job.StatusRunning // Can only complete if running
	updatedAt := j.UpdatedAt

	result := "completed with results"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err = j.Complete(result)

	h.NoError(err)
	h.Equal(job.StatusCompleted, j.Status)
	h.Equal(result, j.Result)
	h.True(j.UpdatedAt.After(updatedAt))
}

func TestComplete_FailsIfNotRunning(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(job.StatusPending, j.Status) // Is in pending state

	err = j.Complete("some result")
	h.ErrorIs(err, job.ErrCannotBeCompleted)
}

func TestMarkAsFailed_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	j.Status = job.StatusRunning // Can only fail if running
	updatedAt := j.UpdatedAt

	errMsg := "transcoding failed"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err = j.MarkAsFailed(errMsg)

	h.NoError(err)
	h.Equal(job.StatusFailed, j.Status)
	h.Equal(errMsg, j.ErrorMsg)
	h.True(j.UpdatedAt.After(updatedAt))
}

func TestMarkAsFailed_FailsIfNotRunning(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	j, err := job.NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(job.StatusPending, j.Status) // Is in pending state

	err = j.MarkAsFailed("some error")
	h.ErrorIs(err, job.ErrCannotBeMarkedAsFailed)
}
