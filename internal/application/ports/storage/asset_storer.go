package storage

import (
	"context"
	"io"
)

type AssetStorer interface {
	// Save stores a new asset
	// `resourceID` is the top level folder (e.g., the video's UUID)
	// `assetPath` is the path within that folder (e.g., "original.mp4", or "transcoded/360.m4s")
	// `content` is the file data to be written
	Save(ctx context.Context, resourceID, assetPath string, content io.Reader) error

	// DeleteAll deletes all the content within the folder specified by the `resourceID`
	DeleteAll(ctx context.Context, resourceID string) error
}
