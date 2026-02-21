package transcode

import "time"

type TranscodeOutput struct {
	Duration     time.Duration
	ManifestPath string // The relative path to the generated manifest file
}
