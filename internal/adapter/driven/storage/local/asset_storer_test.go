package local

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLocalAssetStorer(t *testing.T) {
	t.Parallel()

	// Use t.TempDir() to create a temporary directory.
	tempDir := t.TempDir()

	// Path for a new directory that does not exist yet.
	newBasePath := filepath.Join(tempDir, "new_storage")

	storer, err := NewLocalAssetStorer(newBasePath)
	require.NoError(t, err)
	require.NotNil(t, storer)

	// Verify that the directory was actually created on the filesystem.
	_, err = os.Stat(newBasePath)
	require.NoError(t, err, "expected base path directory to be created")
}

func TestSave(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	tempDir := t.TempDir()
	storer, err := NewLocalAssetStorer(tempDir)
	require.NoError(t, err)

	resourceID := "test-resource-123"
	assetPath := "videos/original.mp4"
	content := "hello world"
	contentReader := strings.NewReader(content)

	// --- ACT ---
	err = storer.Save(t.Context(), resourceID, assetPath, contentReader)

	// --- require ---
	require.NoError(t, err)

	// Verify the side effect: check that the file was created with the correct content.
	expectedPath := filepath.Join(tempDir, resourceID, assetPath)
	savedContent, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "expected file to exist at the specified path")
	require.Equal(t, string(savedContent), content)
}

func TestDeleteAll(t *testing.T) {
	t.Parallel()

	// --- ARRANGE ---
	tempDir := t.TempDir()
	storer, err := NewLocalAssetStorer(tempDir)
	require.NoError(t, err)

	resourceID := "test-resource-to-delete"
	// First, save a file to ensure there is something to delete.
	err = storer.Save(t.Context(), resourceID, "some_file.txt", strings.NewReader("delete me"))
	require.NoError(t, err)

	resourcePath := filepath.Join(tempDir, resourceID)
	_, err = os.Stat(resourcePath)
	require.NoError(t, err, "sanity check: expected resource directory to exist before deletion")

	// --- ACT ---
	err = storer.DeleteAll(t.Context(), resourceID)

	// --- require ---
	require.NoError(t, err)

	// Verify the side effect: check that the directory no longer exists.
	_, err = os.Stat(resourcePath)
	require.True(t, os.IsNotExist(err), "expected resource directory to be deleted")
}
