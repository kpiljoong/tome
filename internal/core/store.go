package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/kpiljoong/tome/internal/util"
)

var errAlreadySaved = errors.New("already saved")

func SaveDir(namespace, root string, smart bool) ([]*model.JournalEntry, error) {
	if err := paths.ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}

	var entries []*model.JournalEntry
	var failures []error
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			failures = append(failures, fmt.Errorf("walk %s: %w", path, err))
			return nil
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		entry, err := Save(namespace, path, smart)
		if errors.Is(err, errAlreadySaved) {
			return nil
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("save %s: %w", path, err))
			return nil
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return entries, err
	}
	if len(failures) > 0 {
		return entries, fmt.Errorf("save directory completed with %d failure(s): %w", len(failures), errors.Join(failures...))
	}
	return entries, nil
}

func SaveDirWithExclude(namespace, root string, smart bool, excludes []string) ([]*model.JournalEntry, error) {
	if err := paths.ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}

	var entries []*model.JournalEntry
	var failures []error

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			failures = append(failures, fmt.Errorf("walk %s: %w", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if util.ShouldExclude(path, excludes) {
			return nil
		}

		entry, err := Save(namespace, path, smart)
		if errors.Is(err, errAlreadySaved) {
			return nil
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("save %s: %w", path, err))
			return nil
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return entries, err
	}
	if len(failures) > 0 {
		return entries, fmt.Errorf("save directory completed with %d failure(s): %w", len(failures), errors.Join(failures...))
	}
	return entries, nil
}

// Save saves a file to the journal under a given namespace.
func Save(namespace, path string, smart bool) (*model.JournalEntry, error) {
	if err := paths.ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve full path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := computeBlobHash(data)

	if smart {
		existing, _ := Search(namespace, filepath.Base(path))
		for _, e := range existing {
			if e.FullPath == absPath && e.BlobHash == hash {
				// Still append metadata if file meta changed (mtime or size)
				info, statErr := os.Stat(absPath)
				if statErr == nil {
					if e.Meta["mtime"] != info.ModTime().Format(time.RFC3339) ||
						e.Meta["size"] != fmt.Sprintf("%d", info.Size()) {
						break // proceed to create new entry
					}
				}
				return nil, errAlreadySaved
			}
		}
	}

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

func computeBlobHash(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
