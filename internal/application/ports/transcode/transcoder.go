package transcode

import "context"

type Transcoder interface {
	// Transcode takes a source video asset, converts it into a streaming format,
	// and places the output into the same resource location.
	// it returns metadata about the transcoded assets
	Transcode(ctx context.Context, resourceID, sourceFilename string) (*TranscodeOutput, error)
}
