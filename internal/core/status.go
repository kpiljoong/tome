package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/pkg/model"
)

type SyncStatus struct {
	ID        string
	Namespace string
	Filename  string
	Source    string // local, remote, synced, conflict
}

func Status(localPath string, remote backend.RemoteBackend, jsonOut bool) error {
	journalRoot := filepath.Join(localPath, "journals")

	localNamespaces, _ := listLocalNamespaces(journalRoot)
	remoteNamespaces, _ := remote.ListNamespaces()

	nsSet := map[string]bool{}
	for _, ns := range append(localNamespaces, remoteNamespaces...) {
		nsSet[ns] = true
	}

	var statuses []SyncStatus

	for ns := range nsSet {
		localEntries := map[string]*model.JournalEntry{}
		remoteEntries := map[string]*model.JournalEntry{}

		// Load local
		files, _ := os.ReadDir(filepath.Join(journalRoot, ns))
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".json") {
				data, _ := os.ReadFile(filepath.Join(journalRoot, ns, f.Name()))
				var entry model.JournalEntry
				if err := json.Unmarshal(data, &entry); err == nil {
					localEntries[entry.ID] = &entry
				}
			}
		}

		// Load remote
		rlist, _ := remote.ListJournal(ns, "")
		for _, r := range rlist {
			remoteEntries[r.ID] = r
		}

		// Compare
		seen := map[string]bool{}
		for id, local := range localEntries {
			if remote, ok := remoteEntries[id]; ok {
				if local.BlobHash == remote.BlobHash {
					statuses = append(statuses, SyncStatus{
						ID:        id,
						Namespace: ns,
						Filename:  local.Filename,
						Source:    "synced",
					})
				} else {
					statuses = append(statuses, SyncStatus{
						ID:        id,
						Namespace: ns,
						Filename:  local.Filename,
						Source:    "conflict",
					})
				}
				seen[id] = true
			} else {
				statuses = append(statuses, SyncStatus{
					ID:        id,
					Namespace: ns,
					Filename:  local.Filename,
					Source:    "local",
				})
			}
		}

		for id, remote := range remoteEntries {
			if seen[id] {
				continue
			}
			statuses = append(statuses, SyncStatus{
				ID:        id,
				Namespace: ns,
				Filename:  remote.Filename,
				Source:    "remote",
			})
		}
	}

	// Output
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(statuses)
	} else {
		for _, s := range statuses {
			var tag string
			switch s.Source {
			case "local":
				tag = "🆕  Only in local:"
			case "remote":
				tag = "☁️  Only in remote:"
			case "synced":
				tag = "✅  Synced:"
			case "conflict":
				tag = "⚠️  Conflict:"
			}
			fmt.Printf("%s  %s/%s\n", tag, s.Namespace, s.Filename)
		}
		return nil
	}
}
