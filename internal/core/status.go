package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/kpiljoong/tome/internal/util"
)

// type SyncStatus struct {
// 	ID        string `json:"id"`
// 	Namespace string `json:"namespace"`
// 	Filename  string `json:"filename"`
// 	Source    string `json:"source"` // local, remote, synced, conflict
// }

func Status(localPath string, remote backend.RemoteBackend, jsonOut bool) error {
	journalRoot := paths.JournalsDir()

	localNamespaces, _ := listLocalNamespaces(journalRoot)
	remoteNamespaces, _ := remote.ListNamespaces()

	nsSet := map[string]bool{}
	for _, ns := range append(localNamespaces, remoteNamespaces...) {
		nsSet[ns] = true
	}

	var statuses []cliutil.SyncStatus

	for ns := range nsSet {
		logx.Info("📂 Checking namespace: %s", ns)

		localEntries, err := loadLocalEntries(paths.NamespaceDir(ns))
		if err != nil {
			logx.Warn("Failed to load local entries for %s: %v", ns, err)
			continue
		}

		remoteEntries, err := loadRemoteEntries(remote, ns)
		if err != nil {
			logx.Warn("Failed to load remote entries for %s: %v", ns, err)
			continue
		}

		util.SortJournalMapByTimestampDesc(localEntries)
		util.SortJournalMapByTimestampDesc(remoteEntries)

		// Compare entries
		seen := map[string]bool{}
		for id, local := range localEntries {
			if remote, ok := remoteEntries[id]; ok {
				if local.BlobHash == remote.BlobHash {
					statuses = append(statuses, cliutil.NewStatus(ns, id, local.Filename, "synced"))
				} else {
					statuses = append(statuses, cliutil.NewStatus(ns, id, local.Filename, "conflict"))
				}
				seen[id] = true
			} else {
				statuses = append(statuses, cliutil.NewStatus(ns, id, local.Filename, "local"))
			}
		}

		for id, remote := range remoteEntries {
			if seen[id] {
				continue
			}
			statuses = append(statuses, cliutil.NewStatus(ns, id, remote.Filename, "remote"))
		}
	}

	// Output
	return cliutil.PrintStatus(statuses, jsonOut)
	//	if jsonOut {
	//		return cliutil.PrintPrettyJSON(statuses)
	//	}
	//
	//	if len(statuses) == 0 {
	//		logx.Success("✅ Everything is in sync")
	//		return nil
	//	}
	//
	// return nil
}

func loadLocalEntries(nsPath string) (map[string]*model.JournalEntry, error) {
	entries := map[string]*model.JournalEntry{}

	files, err := os.ReadDir(nsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read namespace dir: %w", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(nsPath, f.Name()))
		if err != nil {
			continue
		}
		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err == nil {
			entries[entry.ID] = &entry
		}
	}

	return entries, nil
}

func loadRemoteEntries(remote backend.RemoteBackend, ns string) (map[string]*model.JournalEntry, error) {
	entries := map[string]*model.JournalEntry{}

	rlist, err := remote.ListJournal(ns, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote journal: %w", err)
	}

	for _, r := range rlist {
		entries[r.ID] = r
	}

	return entries, nil
}
