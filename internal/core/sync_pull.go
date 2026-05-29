package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/paths"
)

func PullNamespace(namespace string, remote backend.RemoteBackend) error {
	fmt.Printf("📥 Pulling namespace: %s\n", namespace)

	journalRoot := paths.JournalsDir()
	blobRoot := paths.BlobsDir()

	_ = os.MkdirAll(filepath.Join(journalRoot, namespace), 0o755)
	_ = os.MkdirAll(blobRoot, 0o755)

	entries, err := remote.ListJournal(namespace, "")
	if err != nil {
		return fmt.Errorf("failed to list journal for namespace: %s: %w", namespace, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(journalRoot, namespace, entry.ID+".json")
		entryExists := fileExists(entryPath)
		blobPath := paths.BlobPath(entry.BlobHash)
		blobExists := fileExists(blobPath)
		if entryExists && blobExists {
			continue
		}

		if !blobExists {
			blob, err := remote.GetBlobByHash(entry.BlobHash)
			if err != nil {
				fmt.Printf("⚠️  Failed to fetch blob %s: %v\n", entry.BlobHash, err)
				continue
			}
			if err := os.WriteFile(blobPath, blob, 0o644); err != nil {
				fmt.Printf("⚠️  Failed to write blob file %s: %v\n", blobPath, err)
			}
		}

		if !entryExists {
			data, err := json.MarshalIndent(entry, "", "  ")
			if err != nil {
				fmt.Printf("⚠️  Failed to serialize journal entry %s: %v\n", entry.ID, err)
				continue
			}

			if err := os.WriteFile(entryPath, data, 0o644); err != nil {
				fmt.Printf("⚠️  Failed to write journal file %s: %v\n", entryPath, err)
				continue
			}
		}

		fmt.Printf("✅ Pulled: %s/%s\n", namespace, entry.Filename)
	}
	return nil
}

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
			entryExists := fileExists(entryPath)
			blobPath := paths.BlobPath(entry.BlobHash)
			blobExists := fileExists(blobPath)
			if entryExists && blobExists {
				continue
			}

			if !blobExists {
				blob, err := remote.GetBlobByHash(entry.BlobHash)
				if err != nil {
					fmt.Printf("Failed to fetch blob %s: %v\n", entry.BlobHash, err)
					continue
				}
				_ = os.WriteFile(blobPath, blob, 0o644)
			}

			if !entryExists {
				_ = os.MkdirAll(filepath.Dir(entryPath), 0o755)
				data, _ := json.MarshalIndent(entry, "", "  ")
				_ = os.WriteFile(entryPath, data, 0o644)
			}

			fmt.Printf("Pulled: %s/%s\n", ns, entry.Filename)
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
