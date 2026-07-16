package core

import (
	"fmt"
	"os"

	"github.com/kpiljoong/tome/internal/paths"
)

func GetBlobByHash(hash string) ([]byte, error) {
	path, err := paths.SafeBlobPath(hash)
	if err != nil {
		return nil, fmt.Errorf("invalid blob hash: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	return data, nil
}
