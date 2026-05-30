package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

func SyncBidirectional(localPath string, remote backend.RemoteBackend) error {
	fmt.Println("Syncing both ways...")

	localNamespaces, err := listLocalNamespaces(paths.JournalsDir())
	if err != nil {
		return fmt.Errorf("failed to list local namespaces: %w", err)
	}
	remoteNamespaces, err := remote.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to list remote namespaces: %w", err)
	}

	// Union of namespaces
	nsMap := make(map[string]bool)
	for _, ns := range append(localNamespaces, remoteNamespaces...) {
		nsMap[ns] = true
	}

	var failures []error
	for ns := range nsMap {
		fmt.Printf("Syncing namespace: %s\n", ns)

		remoteEntries := map[string]*model.JournalEntry{}

		localEntries, err := loadLocalEntriesForBidirectional(ns)
		if err != nil {
			failures = append(failures, err)
			continue
		}

		// Load remote
		remoteList, err := remote.ListJournal(ns, "")
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: list remote journal: %w", ns, err))
			continue
		}
		for _, e := range remoteList {
			remoteEntries[e.ID] = e
		}

		// Pull missing from remote
		for id, re := range remoteEntries {
			if _, found := localEntries[id]; found {
				continue
			}
			fmt.Printf("Pulling %s\n", re.Filename)
			if err := pullRemoteEntry(ns, re, remote); err != nil {
				fmt.Printf("Failed to pull %s: %v\n", re.Filename, err)
				failures = append(failures, err)
			}
		}

		// Push missing to remote
		for id, le := range localEntries {
			if _, exists := remoteEntries[id]; exists {
				continue
			}
			fmt.Printf("Pushing %s\n", le.Filename)

			journalPath := paths.JournalPath(ns, le.ID)
			blobPath := paths.BlobPath(le.BlobHash)
			if err := remote.UploadFile(blobPath, filepath.ToSlash(paths.RemoteBlobPath(le.BlobHash))); err != nil {
				failures = append(failures, fmt.Errorf("%s/%s: upload blob %s: %w", ns, le.ID, blobPath, err))
				continue
			}
			if err := remote.UploadFile(journalPath, filepath.ToSlash(paths.RemoteJournalPath(ns, le.ID))); err != nil {
				failures = append(failures, fmt.Errorf("%s/%s: upload journal %s: %w", ns, le.ID, journalPath, err))
			}
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("bidirectional sync completed with %d failure(s): %w", len(failures), errors.Join(failures...))
	}

	fmt.Println("Sync complete.")
	return nil
}

func listLocalNamespaces(journalRoot string) ([]string, error) {
	entries, err := os.ReadDir(journalRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []string
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

func loadLocalEntriesForBidirectional(namespace string) (map[string]*model.JournalEntry, error) {
	entries := map[string]*model.JournalEntry{}
	dir := paths.NamespaceDir(namespace)

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, fmt.Errorf("%s: read local namespace: %w", namespace, err)
	}

	var failures []error
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		entryPath := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(entryPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: read local journal %s: %w", namespace, entryPath, err))
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			failures = append(failures, fmt.Errorf("%s: decode local journal %s: %w", namespace, entryPath, err))
			continue
		}
		entries[entry.ID] = &entry
	}

	if len(failures) > 0 {
		return nil, fmt.Errorf("%s: load local entries completed with %d failure(s): %w", namespace, len(failures), errors.Join(failures...))
	}
	return entries, nil
}

func pullRemoteEntry(namespace string, entry *model.JournalEntry, remote backend.RemoteBackend) error {
	if err := paths.EnsureDirExists(paths.BlobsDir()); err != nil {
		return fmt.Errorf("%s/%s: create blob dir: %w", namespace, entry.ID, err)
	}

	blob, err := remote.GetBlobByHash(entry.BlobHash)
	if err != nil {
		return fmt.Errorf("%s/%s: fetch blob %s: %w", namespace, entry.ID, entry.BlobHash, err)
	}
	if err := os.WriteFile(paths.BlobPath(entry.BlobHash), blob, 0o644); err != nil {
		return fmt.Errorf("%s/%s: write blob %s: %w", namespace, entry.ID, paths.BlobPath(entry.BlobHash), err)
	}

	if err := paths.EnsureDirExists(paths.NamespaceDir(namespace)); err != nil {
		return fmt.Errorf("%s/%s: create journal dir: %w", namespace, entry.ID, err)
	}
	journalDelta, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("%s/%s: serialize journal entry: %w", namespace, entry.ID, err)
	}
	if err := os.WriteFile(paths.JournalPath(namespace, entry.ID), journalDelta, 0o644); err != nil {
		return fmt.Errorf("%s/%s: write journal %s: %w", namespace, entry.ID, paths.JournalPath(namespace, entry.ID), err)
	}

	return nil
}
