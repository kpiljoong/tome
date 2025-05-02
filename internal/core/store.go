package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
	"github.com/kpiljoong/tome/pkg/util"
)

// Save saves a file to the journal under a given namespace.
func Save(namespace, path string) (*model.JournalEntry, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve full path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	sum := sha256.Sum256(data)
	hash := "sha256:" + hex.EncodeToString(sum[:])

	if err := paths.EnsureDirExists(paths.BlobsDir()); err != nil {
		return nil, fmt.Errorf("failed to create blobs directory: %w", err)
	}
	if err := paths.EnsureDirExists(paths.NamespaceDir(namespace)); err != nil {
		return nil, fmt.Errorf("failed to create namespace dir: %w", err)
	}

	// Write blob if not exists
	blobPath := paths.BlobPath(hash)
	if _, err := os.Stat(blobPath); os.IsNotExist(err) {
		if err := os.WriteFile(blobPath, data, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write blob file: %w", err)
		}
	}
	entry := &model.JournalEntry{
		ID:        util.GenerateULID(),
		Timestamp: time.Now().UTC(),
		Namespace: namespace,
		Filename:  filepath.Base(path),
		FullPath:  absPath,
		BlobHash:  hash,
		Meta: map[string]string{
			"size":  fmt.Sprintf("%d", len(data)),
			"mtime": util.ModTime(path).Format(time.RFC3339),
		},
	}

	entryPath := paths.JournalPath(namespace, entry.ID)
	entryData, _ := json.MarshalIndent(entry, "", "  ")

	if err := os.WriteFile(entryPath, entryData, 0o644); err != nil {
		return nil, fmt.Errorf("failed to write journal entry: %w", err)
	}
	return entry, nil
}
