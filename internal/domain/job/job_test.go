package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// jobTestHelper encapsulates common test setup and assertions.
type jobTestHelper struct {
	*require.Assertions
	mockID      string
	mockVideoID string
	mockJobType JobType
}

// setupJobTestHelper creates a new helper for a given test.
func setupJobTestHelper(t *testing.T) *jobTestHelper {
	return &jobTestHelper{
		Assertions:  require.New(t),
		mockID:      "mock_job_id",
		mockVideoID: "mock_video_id",
		mockJobType: TypeTranscode,
	}
}

func TestNewJob_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)

	h.NoError(err)
	h.NotNil(job)
	h.Equal(h.mockID, job.ID)
	h.Equal(h.mockVideoID, job.VideoID)
	h.Equal(h.mockJobType, job.Type)
	h.Equal(StatusPending, job.Status)
}

func TestNewJob_FailsOnEmptyID(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob("", h.mockVideoID, h.mockJobType)

	h.Nil(job)
	h.ErrorIs(err, ErrJobIDEmpty)
}

func TestNewJob_FailsOnEmptyVideoID(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, "", h.mockJobType)

	h.Nil(job)
	h.ErrorIs(err, ErrVideoIDEmpty)
}

func TestNewJob_FailsOnInvalidJobType(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, JobType("invalid_type"))

	h.Nil(job)
	h.ErrorIs(err, ErrJobTypeInvalid)
}

func TestStart_SuccessCaseFromPending(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(StatusPending, job.Status)

	err = job.Start()
	h.NoError(err)
	h.Equal(StatusRunning, job.Status)
}

func TestStart_SuccessCaseFromFailed(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	job.Status = StatusFailed // Manually set state for test

	err = job.Start()
	h.NoError(err)
	h.Equal(StatusRunning, job.Status)
}

func TestStart_FailsIfNotPendingOrFailed(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	job.Status = StatusRunning // Set to a non-startable state

	err = job.Start()
	h.ErrorIs(err, ErrCannotBeStarted)
}

func TestComplete_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	job.Status = StatusRunning // Can only complete if running
	updatedAt := job.UpdatedAt

	result := "completed with results"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err = job.Complete(result)

	h.NoError(err)
	h.Equal(StatusCompleted, job.Status)
	h.Equal(result, job.Result)
	h.True(job.UpdatedAt.After(updatedAt))
}

func TestComplete_FailsIfNotRunning(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(StatusPending, job.Status) // Is in pending state

	err = job.Complete("some result")
	h.ErrorIs(err, ErrCannotBeCompleted)
}

func TestMarkAsFailed_SuccessCase(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	job.Status = StatusRunning // Can only fail if running
	updatedAt := job.UpdatedAt

	errMsg := "transcoding failed"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err = job.MarkAsFailed(errMsg)

	h.NoError(err)
	h.Equal(StatusFailed, job.Status)
	h.Equal(errMsg, job.ErrorMsg)
	h.True(job.UpdatedAt.After(updatedAt))
}

func TestMarkAsFailed_FailsIfNotRunning(t *testing.T) {
	t.Parallel()
	h := setupJobTestHelper(t)

	job, err := NewJob(h.mockID, h.mockVideoID, h.mockJobType)
	h.NoError(err)
	h.Equal(StatusPending, job.Status) // Is in pending state

	err = job.MarkAsFailed("some error")
	h.ErrorIs(err, ErrCannotBeMarkedAsFailed)
}
