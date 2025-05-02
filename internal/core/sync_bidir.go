package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

func SyncBidirectional(localPath string, remote backend.RemoteBackend) error {
	fmt.Println("Syncing both ways...")

	// journalRoot := filepath.Join(localPath, "journals")
	// blobRoot := filepath.Join(localPath, "blobs")

	localNamespaces, _ := listLocalNamespaces(paths.JournalsDir())
	remoteNamespaces, _ := remote.ListNamespaces()

	// Union of namespaces
	nsMap := make(map[string]bool)
	for _, ns := range append(localNamespaces, remoteNamespaces...) {
		nsMap[ns] = true
	}

	for ns := range nsMap {
		fmt.Printf("Syncing namespace: %s\n", ns)

		localEntries := map[string]*model.JournalEntry{}
		remoteEntries := map[string]*model.JournalEntry{}

		// Load local
		// localDir := filepath.Join(journalRoot, ns)
		dir := paths.NamespaceDir(ns)
		// _ = os.MkdirAll(localDir, 0o755)
		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(dir, f.Name()))
			var e model.JournalEntry
			if err := json.Unmarshal(data, &e); err == nil {
				localEntries[e.ID] = &e
			}
		}

		// Load remote
		remoteList, _ := remote.ListJournal(ns, "")
		for _, e := range remoteList {
			remoteEntries[e.ID] = e
		}

		// Pull missing from remote
		for id, re := range remoteEntries {
			if _, found := localEntries[id]; found {
				continue
			}
			fmt.Printf("Pulling %s\n", re.Filename)
			blob, err := remote.GetBlobByHash(re.BlobHash)
			if err != nil {
				fmt.Printf("Failed to pull blob: %v\n", err)
				continue
			}

			// Save blob
			_ = paths.EnsureDirExists(paths.BlobsDir())
			_ = os.WriteFile(paths.BlobPath(re.BlobHash), blob, 0o644)
			// Save journal
			_ = paths.EnsureDirExists(paths.NamespaceDir(ns))
			journalDelta, _ := json.MarshalIndent(re, "", "  ")
			_ = os.WriteFile(paths.JournalPath(ns, re.ID), journalDelta, 0o644)
		}

		// Push missing to remote
		for id, le := range localEntries {
			if _, exists := remoteEntries[id]; exists {
				continue
			}
			fmt.Printf("Pushing %s\n", le.Filename)

			journalPath := paths.JournalPath(ns, le.ID)
			blobPath := paths.BlobPath(le.BlobHash)
			_ = remote.UploadFile(journalPath, filepath.ToSlash(paths.RemoteJournalPath(ns, le.ID)))
			_ = remote.UploadFile(blobPath, filepath.ToSlash(paths.RemoteBlobPath(le.BlobHash)))
		}
	}

	fmt.Println("Sync compelte.")
	return nil
}

func listLocalNamespaces(journalRoot string) ([]string, error) {
	entries, err := os.ReadDir(journalRoot)
	if err != nil {
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
