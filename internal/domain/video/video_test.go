package video

import (
	"testing"
	"time"

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

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	h.NoError(err)
	h.NotNil(video)
	h.Equal(video.ID, h.mockID)
	h.Equal(video.Title, h.mockTitle)
	h.Equal(video.Description, h.mockDescription)
	h.Equal(video.ResourceID, h.mockResourceID)
	h.Equal(StatusPending, video.Status)
}

func TestNewVideo_EmptyID(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo("", h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	h.Nil(video)
	h.ErrorIs(err, ErrVideoIDEmpty)
}

func TestNewVideo_EmptyFilename(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, "", h.mockResourceID)

	h.Nil(video)
	h.ErrorIs(err, ErrFilenameEmpty)
}

func TestNewVideo_EmptyResourceID(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, "")

	h.Nil(video)
	h.ErrorIs(err, ErrResourceIDEmpty)
}

func TestMarkAsProcessing_SuccessCaseFromPending(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)

	err = video.MarkAsProcessing()
	h.NoError(err)
	h.Equal(StatusProcessing, video.Status)
}

func TestMarkAsProcessing_SuccessCaseFromFailed(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusFailed // Manually set state for test

	err = video.MarkAsProcessing()
	h.NoError(err)
	h.Equal(StatusProcessing, video.Status)
}

func TestMarkAsProcessing_CannotProcessIfPublished(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusPublished

	err = video.MarkAsProcessing()
	h.ErrorIs(err, ErrCannotBeMarkedAsProcessing)
}

func TestMarkAsProcessing_CannotProcessIfArchived(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusArchived

	err = video.MarkAsProcessing()
	h.ErrorIs(err, ErrCannotBeMarkedAsProcessing)
}

func TestMarkAsFailed_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusProcessing // Can only fail if processing

	err = video.MarkAsFailed()
	h.NoError(err)
	h.Equal(StatusFailed, video.Status)
}

func TestMarkAsFailed_CannotFailIfNotProcessing(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(StatusPending, video.Status) // Starts as pending

	err = video.MarkAsFailed()
	h.ErrorIs(err, ErrCannotBeMarkedAsFailed)
}

func TestPublish_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusProcessing // Can only publish if processing

	err = video.Publish()
	h.NoError(err)
	h.Equal(StatusPublished, video.Status)
}

func TestPublish_CannotPublishIfNotProcessing(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(StatusPending, video.Status) // Starts as pending

	err = video.Publish()
	h.ErrorIs(err, ErrCannotBePublished)
}

func TestArchive_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	video.Status = StatusPublished // Can only archive if published

	err = video.Archive()
	h.NoError(err)
	h.Equal(StatusArchived, video.Status)
}

func TestArchive_CannotArchiveIfNotPublished(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)

	video, err := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	h.NoError(err)
	h.Equal(StatusPending, video.Status) // Starts as pending

	err = video.Archive()
	h.ErrorIs(err, ErrCannotBeArchived)
}

func TestUpdateTitle_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	updatedAt := video.UpdatedAt

	newTitle := "new_title"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err := video.UpdateTitle(newTitle)

	h.NoError(err)
	h.Equal(newTitle, video.Title)
	h.True(video.UpdatedAt.After(updatedAt))
}

func TestUpdateTitle_FailsOnEmptyTitle(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := video.UpdateTitle("")
	h.ErrorIs(err, ErrTitleEmpty)
}

func TestUpdateDescription_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	updatedAt := video.UpdatedAt

	newDescription := "new_description"
	time.Sleep(1 * time.Nanosecond) // Ensure UpdatedAt changes
	err := video.UpdateDescription(newDescription)

	h.NoError(err)
	h.Equal(newDescription, video.Description)
	h.True(video.UpdatedAt.After(updatedAt))
}

func TestUpdateDescription_FailsOnEmptyDescription(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := video.UpdateDescription("")
	h.ErrorIs(err, ErrDescriptionEmpty)
}

func TestUpdateDuration_SuccessCase(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	duration := 120 * time.Second
	err := video.UpdateDuration(duration)

	h.NoError(err)
	h.Equal(duration, video.Duration)
}

func TestUpdateDuration_FailsIfAlreadySet(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)
	video.Duration = 100 * time.Second // Pre-set duration

	err := video.UpdateDuration(120 * time.Second)
	h.ErrorIs(err, ErrDurationAlreadySet)
}

func TestUpdateDuration_FailsOnNegativeDuration(t *testing.T) {
	t.Parallel()

	h := setupVideoTestHelper(t)
	video, _ := NewVideo(h.mockID, h.mockTitle, h.mockDescription, h.mockFilename, h.mockResourceID)

	err := video.UpdateDuration(-10 * time.Second)
	h.ErrorIs(err, ErrDurationNegative)
}
