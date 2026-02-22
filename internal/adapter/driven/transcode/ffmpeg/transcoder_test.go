package ffmpeg

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	execmocks "github.com/st-ember/streaming-api/internal/application/ports/exec/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDuration_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)

	// Mock the JSON output from a successful ffprobe
	ffprobeOutput := `{"format":{"duration":"123.45"}}`

	// Expectations
	mockCommander.EXPECT().
		CommandContext(mock.Anything, "ffprobe", mock.Anything).
		Return(mockCmd).
		Once()

	// Expect SetStdout to be called, and we can write our fake JSON output to it.
	mockCmd.EXPECT().SetStdout(mock.Anything).Run(func(w io.Writer) {
		w.Write([]byte(ffprobeOutput))
	}).Once()

	mockCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockCmd.EXPECT().Run().Return(nil).Once()

	// --- ACT ---
	transcoder := NewFFMPEGTranscoder("/tmp", mockCommander)
	duration, err := transcoder.getDuration(context.Background(), "/tmp/some/path.mp4")

	// --- ASSERT ---
	require.NoError(t, err)
	// 123.45 seconds should be correctly parsed.
	expectedDuration := time.Duration(123.45 * float64(time.Second))
	require.Equal(t, expectedDuration, duration)
}

func TestGetDuration_FailsOnCommandRun(t *testing.T) {
	t.Parallel()
	mockCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)
	expectedErr := errors.New("ffprobe command not found")

	mockCommander.EXPECT().CommandContext(mock.Anything, "ffprobe", mock.Anything).Return(mockCmd).Once()
	mockCmd.EXPECT().SetStdout(mock.Anything).Once()
	mockCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockCmd.EXPECT().Run().Return(expectedErr).Once() // Simulate ffprobe failing to run

	transcoder := NewFFMPEGTranscoder("/tmp", mockCommander)
	_, err := transcoder.getDuration(context.Background(), "/tmp/some/path.mp4")

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestTranscode_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockProbeCmd := execmocks.NewMockCmd(t)
	mockFFmpegCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)

	// ffprobe setup
	ffprobeOutput := `{"format":{"duration":"120.0"}}`
	mockCommander.EXPECT().
		CommandContext(mock.Anything, "ffprobe", mock.Anything).
		Return(mockProbeCmd).
		Once()
	mockProbeCmd.EXPECT().SetStdout(mock.Anything).Run(func(w io.Writer) { w.Write([]byte(ffprobeOutput)) }).Once()
	mockProbeCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockProbeCmd.EXPECT().Run().Return(nil).Once()

	// ffmpeg setup
	mockCommander.EXPECT().
		CommandContext(mock.Anything, "ffmpeg", mock.Anything).
		Return(mockFFmpegCmd).
		Once()
	mockFFmpegCmd.EXPECT().SetStderr(mock.Anything).Once()
	mockFFmpegCmd.EXPECT().Run().Return(nil).Once() // ffmpeg succeeds

	// --- ACT ---
	transcoder := NewFFMPEGTranscoder("/tmp", mockCommander)
	// We need to create a temporary source file for ffprobe to not fail on missing file
	tmpFile, err := os.CreateTemp("", "source-*.mp4")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// For this test, we can pass dummy values for resourceID and filename,
	// as the mock commander doesn't actually use them.
	output, err := transcoder.Transcode(context.Background(), "resource-id", tmpFile.Name())

	// --- ASSERT ---
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, 120*time.Second, output.Duration)
	require.Contains(t, output.ManifestPath, "manifest.mpd")

	// Clean up the temporary directory created by the function
	if output != nil {
		os.RemoveAll(filepath.Dir(output.ManifestPath))
	}
}

func TestTranscode_FailsOnGetDuration(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockProbeCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)
	expectedErr := errors.New("ffprobe failed")

	// ffprobe fails
	mockCommander.EXPECT().CommandContext(mock.Anything, "ffprobe", mock.Anything).Return(mockProbeCmd).Once()
	mockProbeCmd.EXPECT().SetStdout(mock.Anything).Once()
	mockProbeCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockProbeCmd.EXPECT().Run().Return(expectedErr).Once()

	// --- ACT ---
	transcoder := NewFFMPEGTranscoder("/tmp", mockCommander)
	_, err := transcoder.Transcode(context.Background(), "resource-id", "source.mp4")

	// --- ASSERT ---
	require.Error(t, err)
	require.ErrorContains(t, err, "get duration")
	require.ErrorIs(t, err, expectedErr)
}

func TestTranscode_FailsOnFFmpegRun(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockProbeCmd := execmocks.NewMockCmd(t)
	mockFFmpegCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)

	// ffprobe setup (succeeds)
	ffprobeOutput := `{"format":{"duration":"120.0"}}`
	mockCommander.EXPECT().CommandContext(mock.Anything, "ffprobe", mock.Anything).Return(mockProbeCmd).Once()
	mockProbeCmd.EXPECT().SetStdout(mock.Anything).Run(func(w io.Writer) { w.Write([]byte(ffprobeOutput)) }).Once()
	mockProbeCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockProbeCmd.EXPECT().Run().Return(nil).Once()

	// ffmpeg setup (fails)
	expectedErr := errors.New("invalid codec")
	mockCommander.EXPECT().CommandContext(mock.Anything, "ffmpeg", mock.Anything).Return(mockFFmpegCmd).Once()

	// When SetStderr is called, we can write a fake error message to the buffer
	mockFFmpegCmd.EXPECT().SetStderr(mock.Anything).Run(func(w io.Writer) {
		w.Write([]byte("ffmpeg error: something went wrong"))
	}).Once()

	mockFFmpegCmd.EXPECT().Run().Return(expectedErr).Once() // ffmpeg fails

	// --- ACT ---
	transcoder := NewFFMPEGTranscoder("/tmp", mockCommander)
	tmpFile, err := os.CreateTemp("", "source-*.mp4")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = transcoder.Transcode(context.Background(), "resource-id", tmpFile.Name())

	// --- ASSERT ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "ffmpeg error: something went wrong")
}
