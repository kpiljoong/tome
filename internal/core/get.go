package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Get(namespace, query string) ([]byte, error) {
	entries, err := Search(namespace, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries found for query: %s", query)
	}

	if len(entries) > 1 {
		fmt.Println("Multiple matches found:")
		for _, e := range entries {
			fmt.Printf("  - [%s] %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.FullPath)
		}
		return nil, errors.New("ambigous result - refine your query")
	}

	entry := entries[0]
	return readBlob(entry.BlobHash)
}

func readBlob(hash string) ([]byte, error) {
	if !strings.HasPrefix(hash, "sha256:") {
		return nil, fmt.Errorf("invalid hash format: %s", hash)
	}

	baseDir := filepath.Join(os.Getenv("HOME"), ".tome", "blobs")
	blobPath := filepath.Join(baseDir, hash)

	data, err := os.ReadFile(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob file: %w", err)
	}
	return data, nil
}
