package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetBlobByHash(hash string) ([]byte, error) {
	baseDir := filepath.Join(os.Getenv("HOME"), ".tome", "blobs")
	blobPath := filepath.Join(baseDir, hash)

	data, err := os.ReadFile(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	return data, nil
}
