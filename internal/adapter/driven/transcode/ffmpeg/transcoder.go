package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"path/filepath"
	"strconv"
	"time"

	"github.com/st-ember/streaming-api/internal/application/ports/exec"
	"github.com/st-ember/streaming-api/internal/application/ports/transcode"
)

type FFMPEGTranscoder struct {
	basePath  string
	commander exec.Commander
}

func NewFFMPEGTranscoder(basePath string, commander exec.Commander) *FFMPEGTranscoder {
	return &FFMPEGTranscoder{basePath, commander}
}

// probeResult is used to unmarshal the json result from ffprobe
type probeResult struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// getDuration gets the duration of a video file in seconds.
func (t *FFMPEGTranscoder) getDuration(ctx context.Context, sourcePath string) (time.Duration, error) {
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", sourcePath,
	}

	cmd := t.commander.CommandContext(ctx, "ffprobe", args...)
	var out bytes.Buffer
	cmd.SetStdout(&out)      // Pipe to out var for access
	cmd.SetStderr(os.Stderr) // Pipe ffprobe errors to standard error for visibility

	// Run ffprobe
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("run ffprobe: %w", err)
	}

	// Unmarshal result
	var result probeResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return 0, fmt.Errorf("parse ffprobe output: %w", err)
	}

	// Convert to float
	durationFloat, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration from ffprobe output: %w", err)
	}

	// Convert to duration and return
	return time.Duration(durationFloat * float64(time.Second)), nil
}

func (t *FFMPEGTranscoder) Transcode(ctx context.Context, resourceID, sourceFilename string) (*transcode.TranscodeOutput, error) {
	// Assemble full path
	sourcePath := filepath.Join(t.basePath, resourceID, sourceFilename)

	// Get duration
	duration, err := t.getDuration(ctx, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("get duration: %w", err)
	}

	// Create temp dir for transcode output
	// The worker will move the files for permanent storage
	outputDir, err := os.MkdirTemp("", "transcode-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary directory for output: %w", err)
	}

	// TODO: worker will need to delete the temp dir

	manifestPath := filepath.Join(outputDir, "manifest.mpd")

	args := []string{
		"-i", sourcePath,
		"-c:a", "aac", "-ac", "2",
		"-map", "0:v:0", "-map", "0:v:0", "-map", "0:a:0",
		"-c:v:0", "libx264", "-crf:v:0", "23", "-preset:v:0", "medium", "-maxrate:v:0", "1500k",
		"-bufsize:v:0", "3000k", "-s:v:0", "854x480",
		"-c:v:1", "libx264", "-crf:v:1", "22", "-preset:v:1", "medium", "-maxrate:v:1", "3000k",
		"-bufsize:v:1", "6000k", "-s:v:1", "1280x720",
		"-adaptation_sets", "id=0,streams=v id=1,streams=a",
		"-f", "dash",
		manifestPath,
	}

	// Build command
	cmd := t.commander.CommandContext(ctx, "ffmpeg", args...)

	// Capture standard errors to track progress and error details from ffmpeg
	var stdErr bytes.Buffer
	cmd.SetStderr(&stdErr)

	// Execute command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg execution: %w\noutput:\n%s", err, stdErr.String())
	}

	return &transcode.TranscodeOutput{
		Duration:     duration,
		ManifestPath: manifestPath,
	}, nil
}
