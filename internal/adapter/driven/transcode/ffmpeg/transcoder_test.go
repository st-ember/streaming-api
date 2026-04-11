package ffmpeg_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/st-ember/streaming-api/internal/adapter/driven/transcode/ffmpeg"
	execmocks "github.com/st-ember/streaming-api/internal/application/ports/exec/mocks"
	logmocks "github.com/st-ember/streaming-api/internal/application/ports/log/mocks"
	streamermocks "github.com/st-ember/streaming-api/internal/application/ports/progressstream/mocks"
	"github.com/st-ember/streaming-api/internal/domain/progress"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDuration_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	// Mock the JSON output from a successful ffprobe
	ffprobeOutput := `{"format":{"duration":"123.45"}, "streams":{"nb_read_frames":"2962"}}`

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
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	duration, frames, err := transcoder.GetDuration(t.Context(), "/tmp/some/path.mp4")

	// --- ASSERT ---
	require.NoError(t, err)
	// 123.45 seconds should be correctly parsed.
	expectedDuration := time.Duration(123.45 * float64(time.Second))
	require.Equal(t, expectedDuration, duration)
	require.Equal(t, int64(2962), frames)
}

func TestGetDuration_FailsOnCommandRun(t *testing.T) {
	t.Parallel()
	mockCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)
	expectedErr := errors.New("ffprobe command not found")

	mockCommander.EXPECT().CommandContext(mock.Anything, "ffprobe", mock.Anything).Return(mockCmd).Once()
	mockCmd.EXPECT().SetStdout(mock.Anything).Once()
	mockCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockCmd.EXPECT().Run().Return(expectedErr).Once() // Simulate ffprobe failing to run

	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	_, _, err := transcoder.GetDuration(t.Context(), "/tmp/some/path.mp4")

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
}

func TestPipeProgress_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	jobID := "test-job-id"
	totalFrames := int64(300)

	// Simulate FFmpeg's "-progress pipe:1" output
	progressData := "frame=100\nframe=200\nframe=300\n"
	progressPipe := io.NopCloser(strings.NewReader(progressData))

	// Expect calls for each frame update line (Status: "continue")
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 100 && p.Status == progress.StatusContinue
	})).Return(nil).Once()

	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 200 && p.Status == progress.StatusContinue
	})).Return(nil).Once()

	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 300 && p.Status == progress.StatusContinue
	})).Return(nil).Once()

	// Expect the final push after the pipe is closed (Status: "end")
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 300 && p.Status == progress.StatusEnd
	})).Return(nil).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	transcoder.PipeProgress(t.Context(), jobID, totalFrames, progressPipe)
}

func TestPipeProgress_ScannerError(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	jobID := "test-job-id"
	totalFrames := int64(300)

	// Simulate an error during reading the pipe
	errReader, errWriter := io.Pipe()
	go func() {
		errWriter.Write([]byte("frame=100\n"))
		errWriter.CloseWithError(errors.New("reading error"))
	}()

	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 100 && p.Status == progress.StatusContinue
	})).Return(nil).Once()

	// Final push should have "error" status
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.Status == progress.StatusError
	})).Return(nil).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	transcoder.PipeProgress(t.Context(), jobID, totalFrames, errReader)
}

func TestPipeProgress_InvalidFrameFormat(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	jobID := "test-job-id"
	totalFrames := int64(300)

	// One invalid frame line, followed by a valid one
	progressData := "frame=abc\nframe=100\n"
	progressPipe := io.NopCloser(strings.NewReader(progressData))

	// Should log error for "frame=abc"
	mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

	// Should process "frame=100" normally
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.CurrentFrames == 100 && p.Status == progress.StatusContinue
	})).Return(nil).Once()

	// Final push with status "end"
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.MatchedBy(func(p *progress.Progress) bool {
		return p.Status == progress.StatusEnd
	})).Return(nil).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	transcoder.PipeProgress(t.Context(), jobID, totalFrames, progressPipe)
}

func TestPipeProgress_NewProgressError(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	jobID := "test-job-id"
	// NewProgress fails if totalFrames is 0
	totalFrames := int64(0)
	progressPipe := io.NopCloser(strings.NewReader("frame=100\n"))

	// Should log error and return immediately
	mockLogger.EXPECT().Errorf(mock.Anything, "start new progress: %v", mock.Anything).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	transcoder.PipeProgress(t.Context(), jobID, totalFrames, progressPipe)
}

func TestPipeProgress_PushError(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	jobID := "test-job-id"
	totalFrames := int64(300)
	progressPipe := io.NopCloser(strings.NewReader("frame=100\n"))

	// Simulate push failing
	expectedErr := errors.New("push error")
	mockStreamer.EXPECT().Push(mock.Anything, jobID, mock.Anything).Return(expectedErr).Once()

	// Should log error and return
	mockLogger.EXPECT().Errorf(mock.Anything, mock.Anything, mock.Anything).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	transcoder.PipeProgress(t.Context(), jobID, totalFrames, progressPipe)
}

func TestTranscode_SuccessCase(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	mockProbeCmd := execmocks.NewMockCmd(t)
	mockFFmpegCmd := execmocks.NewMockCmd(t)
	mockCommander := execmocks.NewMockCommander(t)
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	// ffprobe setup
	ffprobeOutput := `{"format":{"duration":"120.0"}, "streams":{"nb_read_frames":"1000"}}`
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
	mockFFmpegCmd.EXPECT().StdoutPipe().Return(io.NopCloser(strings.NewReader("")), nil).Once()
	mockFFmpegCmd.EXPECT().Start().Return(nil).Once()
	mockFFmpegCmd.EXPECT().Wait().Return(nil).Once()

	// streamer setup
	mockStreamer.EXPECT().Push(mock.Anything, "job-id", mock.Anything).Return(nil)

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	// We need to create a temporary source file for ffprobe to not fail on missing file
	tmpFile, err := os.CreateTemp("", "source-*.mp4")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// For this test, we can pass dummy values for resourceID and filename,
	// as the mock commander doesn't actually use them.
	output, err := transcoder.Transcode(t.Context(), "resource-id", tmpFile.Name(), "job-id")

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
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)
	expectedErr := errors.New("ffprobe failed")

	// ffprobe fails
	mockCommander.EXPECT().CommandContext(mock.Anything, "ffprobe", mock.Anything).Return(mockProbeCmd).Once()
	mockProbeCmd.EXPECT().SetStdout(mock.Anything).Once()
	mockProbeCmd.EXPECT().SetStderr(os.Stderr).Once()
	mockProbeCmd.EXPECT().Run().Return(expectedErr).Once()

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	_, err := transcoder.Transcode(t.Context(), "resource-id", "source.mp4", "job-id")

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
	mockStreamer := streamermocks.NewMockProgressStreamer(t)
	mockLogger := logmocks.NewMockLogger(t)

	// ffprobe setup (succeeds)
	ffprobeOutput := `{"format":{"duration":"123.45"}, "streams":{"nb_read_frames":"2962"}}`
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
	mockFFmpegCmd.EXPECT().StdoutPipe().Return(io.NopCloser(strings.NewReader("")), nil).Once()
	mockFFmpegCmd.EXPECT().Start().Return(expectedErr).Once() // ffmpeg fails

	// --- ACT ---
	transcoder := ffmpeg.NewFFMPEGTranscoder("/tmp", mockCommander, mockStreamer, mockLogger)
	tmpFile, err := os.CreateTemp("", "source-*.mp4")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = transcoder.Transcode(t.Context(), "resource-id", tmpFile.Name(), "job-id")

	// --- ASSERT ---
	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "ffmpeg error: something went wrong")
}
