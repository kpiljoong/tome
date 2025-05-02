package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/pkg/paths"
)

func Pull(localPath string, remote backend.RemoteBackend) error {
	fmt.Println("Pulling from remote...")

	// journalRoot := filepath.Join(localPath, "journals")
	journalRoot := paths.JournalsDir()
	// blobRoot := filepath.Join(localPath, "blobs")
	blobRoot := paths.BlobsDir()

	_ = os.MkdirAll(journalRoot, 0o755)
	_ = os.MkdirAll(blobRoot, 0o755)

	namespaces, err := remote.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, ns := range namespaces {
		entries, err := remote.ListJournal(ns, "")
		if err != nil {
			fmt.Printf("Skipping namespace %s: %v\n", ns, err)
			continue
		}

		for _, entry := range entries {
			entryPath := filepath.Join(journalRoot, ns, entry.ID+".json")
			if _, err := os.Stat(entryPath); err == nil {
				continue
			}

			// blobPath := filepath.Join(blobRoot, entry.BlobHash)
			blobPath := paths.BlobPath(entry.BlobHash)
			if _, err := os.Stat(blobPath); os.IsNotExist(err) {
				blob, err := remote.GetBlobByHash(entry.BlobHash)
				if err != nil {
					fmt.Printf("Failed to fetch blob %s: %v\n", entry.BlobHash, err)
					continue
				}
				_ = os.WriteFile(blobPath, blob, 0o644)
			}

			_ = os.MkdirAll(filepath.Dir(entryPath), 0o755)
			data, _ := json.MarshalIndent(entry, "", "  ")
			_ = os.WriteFile(entryPath, data, 0o644)

			fmt.Printf("Pulled: %s/%s\n", ns, entry.Filename)
		}
	}
	return nil
}
