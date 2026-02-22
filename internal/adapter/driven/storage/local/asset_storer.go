package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/st-ember/streaming-api/internal/application/ports/storage"
)

type LocalAssetStorer struct {
	basePath string
}

const permissionSet os.FileMode = 0755

// NewLocalAssetStorer creates a new LocalAssetStorer and ensures the base directory exists
func NewLocalAssetStorer(basePath string) (storage.AssetStorer, error) {
	if err := os.MkdirAll(basePath, permissionSet); err != nil {
		return nil, fmt.Errorf("create base storage directory: %w", err)
	}

	return &LocalAssetStorer{basePath}, nil
}

// Save stores a new asset
// `resourceID` is the top level folder (e.g., the video's UUID)
// `assetPath` is the path within that folder (e.g., "original.mp4", or "transcoded/360.m4s")
// `content` is the file data to be written
func (s *LocalAssetStorer) Save(ctx context.Context, resourceID, assetPath string, content io.Reader) error {
	// Assemble full file path
	fullPath := filepath.Join(s.basePath, resourceID, assetPath)

	// Create asset directory
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, permissionSet); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content into destination file
	if _, err := io.Copy(dstFile, content); err != nil {
		return fmt.Errorf("copy content into destination file: %w", err)
	}

	return nil
}

// DeleteAll deletes all the content within the folder specified by the `resourceID`
func (s *LocalAssetStorer) DeleteAll(ctx context.Context, resourceID string) error {
	// Assemble resource root path
	resourcePath := filepath.Join(s.basePath, resourceID)

	if err := os.RemoveAll(resourcePath); err != nil {
		return fmt.Errorf("delete assets for resource %s: %w", resourceID, err)
	}

	return nil
}
