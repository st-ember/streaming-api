package video_test

import (
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/domain/video"
	"github.com/stretchr/testify/require"
)

// videoTestHelper encapsulates common test setup and assertions.
type videoTestHelper struct {
	*require.Assertions
	mockID          string
	mockTitle       string
	mockDescription string
	mockFilename    string
	mockResourceID  string
}

// setupVideoTestHelper creates a new helper for a given test.
func setupVideoTestHelper(t *testing.T) *videoTestHelper {
	return &videoTestHelper{
		Assertions:      require.New(t),
		mockID:          "mock_id",
		mockTitle:       "mock_title",
		mockDescription: "mock_description",
		mockFilename:    "mock_filename",
		mockResourceID:  "mock_resource_id",
	}
}

func TestNewVideo_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	h.NoError(err)
	h.NotNil(v)
	h.Equal(v.ID, h.mockID)
	h.Equal(v.Title, h.mockTitle)
	h.Equal(v.Description, h.mockDescription)
	h.Equal(v.ResourceID, h.mockResourceID)
	h.Equal(video.StatusPending, v.Status)
}

func TestNewVideo_EmptyID(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo("", h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	h.Nil(v)
	h.ErrorIs(err, video.ErrVideoIDEmpty)
}

func TestNewVideo_EmptyFilename(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, "", h.mockResourceID)

	h.Nil(v)
	h.ErrorIs(err, video.ErrFilenameEmpty)
}

func TestNewVideo_EmptyResourceID(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, "")

	h.Nil(v)
	h.ErrorIs(err, video.ErrResourceIDEmpty)
}

func TestMarkAsProcessing_SuccessCaseFromPending(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)

	err = v.MarkAsProcessing()
	h.NoError(err)
	h.Equal(video.StatusProcessing, v.Status)
}

func TestMarkAsProcessing_SuccessCaseFromFailed(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusFailed // Manually set state for test

	err = v.MarkAsProcessing()
	h.NoError(err)
	h.Equal(video.StatusProcessing, v.Status)
}

func TestMarkAsProcessing_CannotProcessIfPublished(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusPublished

	err = v.MarkAsProcessing()
	h.ErrorIs(err, video.ErrCannotBeMarkedAsProcessing)
}

func TestMarkAsProcessing_CannotProcessIfArchived(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusArchived

	err = v.MarkAsProcessing()
	h.ErrorIs(err, video.ErrCannotBeMarkedAsProcessing)
}

func TestMarkAsFailed_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusProcessing // Can only fail if processing

	err = v.MarkAsFailed()
	h.NoError(err)
	h.Equal(video.StatusFailed, v.Status)
}

func TestMarkAsFailed_CannotFailIfNotProcessing(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(video.StatusPending, v.Status) // Starts as pending

	err = v.MarkAsFailed()
	h.ErrorIs(err, video.ErrCannotBeMarkedAsFailed)
}

func TestPublish_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusProcessing // Can only publish if processing

	err = v.Publish()
	h.NoError(err)
	h.Equal(video.StatusPublished, v.Status)
}

func TestPublish_CannotPublishIfNotProcessing(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(video.StatusPending, v.Status) // Starts as pending

	err = v.Publish()
	h.ErrorIs(err, video.ErrCannotBePublished)
}

func TestArchive_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	v.Status = video.StatusPublished // Can only archive if published

	err = v.Archive()
	h.NoError(err)
	h.Equal(video.StatusArchived, v.Status)
}

func TestArchive_CannotArchiveIfNotPublished(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	v, err := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(video.StatusPending, v.Status) // Starts as pending

	err = v.Archive()
	h.ErrorIs(err, video.ErrCannotBeArchived)
}

func TestUpdateTitle_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	updatedAt := v.UpdatedAt

	newTitle := "new_title"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err := v.UpdateTitle(newTitle)

	h.NoError(err)
	h.Equal(newTitle, v.Title)
	h.True(v.UpdatedAt.After(updatedAt))
}

func TestUpdateTitle_FailsOnEmptyTitle(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := v.UpdateTitle("")
	h.ErrorIs(err, video.ErrTitleEmpty)
}

func TestUpdateDescription_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	updatedAt := v.UpdatedAt

	newDescription := "new_description"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err := v.UpdateDescription(newDescription)

	h.NoError(err)
	h.Equal(newDescription, v.Description)
	h.True(v.UpdatedAt.After(updatedAt))
}

func TestUpdateDescription_FailsOnEmptyDescription(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := v.UpdateDescription("")
	h.ErrorIs(err, video.ErrDescriptionEmpty)
}

func TestUpdateDuration_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	duration := 120 * time.Second
	err := v.UpdateDuration(duration)

	h.NoError(err)
	h.Equal(duration, v.Duration)
}

func TestUpdateDuration_FailsIfAlreadySet(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	v.Duration = 100 * time.Second // Pre-set duration

	err := v.UpdateDuration(120 * time.Second)
	h.ErrorIs(err, video.ErrDurationAlreadySet)
}

func TestUpdateDuration_FailsOnNegativeDuration(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	v, _ := video.NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := v.UpdateDuration(-10 * time.Second)
	h.ErrorIs(err, video.ErrDurationNegative)
}
