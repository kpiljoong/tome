package core

import (
	"fmt"
	"os"

	"github.com/kpiljoong/tome/internal/paths"
)

func GetBlobByHash(hash string) ([]byte, error) {
	data, err := os.ReadFile(paths.BlobPath(hash))
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	return data, nil
}
