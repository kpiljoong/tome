package core

import (
	"encoding/json"
	"errors"
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

	var failures []error
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
				failures = append(failures, fmt.Errorf("%s/%s: fetch blob %s: %w", namespace, entry.ID, entry.BlobHash, err))
				continue
			}
			if err := os.WriteFile(blobPath, blob, 0o644); err != nil {
				fmt.Printf("⚠️  Failed to write blob file %s: %v\n", blobPath, err)
				failures = append(failures, fmt.Errorf("%s/%s: write blob %s: %w", namespace, entry.ID, blobPath, err))
				continue
			}
		}

		if !entryExists {
			data, err := json.MarshalIndent(entry, "", "  ")
			if err != nil {
				fmt.Printf("⚠️  Failed to serialize journal entry %s: %v\n", entry.ID, err)
				failures = append(failures, fmt.Errorf("%s/%s: serialize journal entry: %w", namespace, entry.ID, err))
				continue
			}

			if err := os.WriteFile(entryPath, data, 0o644); err != nil {
				fmt.Printf("⚠️  Failed to write journal file %s: %v\n", entryPath, err)
				failures = append(failures, fmt.Errorf("%s/%s: write journal %s: %w", namespace, entry.ID, entryPath, err))
				continue
			}
		}

		fmt.Printf("✅ Pulled: %s/%s\n", namespace, entry.Filename)
	}
	if len(failures) > 0 {
		return fmt.Errorf("pull namespace %s completed with %d failure(s): %w", namespace, len(failures), errors.Join(failures...))
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

	var failures []error
	for _, ns := range namespaces {
		entries, err := remote.ListJournal(ns, "")
		if err != nil {
			fmt.Printf("Skipping namespace %s: %v\n", ns, err)
			failures = append(failures, fmt.Errorf("%s: list journal: %w", ns, err))
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
					failures = append(failures, fmt.Errorf("%s/%s: fetch blob %s: %w", ns, entry.ID, entry.BlobHash, err))
					continue
				}
				if err := os.WriteFile(blobPath, blob, 0o644); err != nil {
					failures = append(failures, fmt.Errorf("%s/%s: write blob %s: %w", ns, entry.ID, blobPath, err))
					continue
				}
			}

			if !entryExists {
				if err := os.MkdirAll(filepath.Dir(entryPath), 0o755); err != nil {
					failures = append(failures, fmt.Errorf("%s/%s: create journal dir: %w", ns, entry.ID, err))
					continue
				}
				data, err := json.MarshalIndent(entry, "", "  ")
				if err != nil {
					failures = append(failures, fmt.Errorf("%s/%s: serialize journal entry: %w", ns, entry.ID, err))
					continue
				}
				if err := os.WriteFile(entryPath, data, 0o644); err != nil {
					failures = append(failures, fmt.Errorf("%s/%s: write journal %s: %w", ns, entry.ID, entryPath, err))
					continue
				}
			}

			fmt.Printf("Pulled: %s/%s\n", ns, entry.Filename)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("pull completed with %d failure(s): %w", len(failures), errors.Join(failures...))
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
