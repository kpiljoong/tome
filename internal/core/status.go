package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

// type SyncStatus struct {
// 	ID        string `json:"id"`
// 	Namespace string `json:"namespace"`
// 	Filename  string `json:"filename"`
// 	Source    string `json:"source"` // local, remote, synced, conflict
// }

func Status(localPath string, remote backend.RemoteBackend, jsonOut bool) error {
	journalRoot := paths.JournalsDir()

	localNamespaces, err := listLocalNamespaces(journalRoot)
	if err != nil {
		return fmt.Errorf("failed to list local namespaces: %w", err)
	}
	remoteNamespaces, err := remote.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to list remote namespaces: %w", err)
	}

	nsSet := map[string]bool{}
	for _, ns := range append(localNamespaces, remoteNamespaces...) {
		nsSet[ns] = true
	}

	var namespaces []string
	for ns := range nsSet {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	var statuses []cliutil.SyncStatus
	var failures []error
	for _, ns := range namespaces {
		logx.Info("📂 Checking namespace: %s", ns)
		nsPath, err := paths.SafeNamespaceDir(ns)
		if err != nil {
			failures = append(failures, fmt.Errorf("validate namespace %s: %w", ns, err))
			continue
		}

		localEntries, err := loadLocalEntries(nsPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("load local entries for %s: %w", ns, err))
			continue
		}

		remoteEntries, err := loadRemoteEntries(remote, ns)
		if err != nil {
			failures = append(failures, fmt.Errorf("load remote entries for %s: %w", ns, err))
			continue
		}

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
	if len(failures) > 0 {
		return fmt.Errorf("status completed with %d failure(s): %w", len(failures), errors.Join(failures...))
	}
	sort.Slice(statuses, func(i, j int) bool {
		if statuses[i].Namespace != statuses[j].Namespace {
			return statuses[i].Namespace < statuses[j].Namespace
		}
		if statuses[i].Filename != statuses[j].Filename {
			return statuses[i].Filename < statuses[j].Filename
		}
		return statuses[i].ID < statuses[j].ID
	})

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
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, fmt.Errorf("failed to read namespace dir: %w", err)
	}

	var failures []error
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(nsPath, f.Name()))
		if err != nil {
			failures = append(failures, fmt.Errorf("read journal %s: %w", f.Name(), err))
			continue
		}
		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			failures = append(failures, fmt.Errorf("decode journal %s: %w", f.Name(), err))
			continue
		}
		entries[entry.ID] = &entry
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("load local entries completed with %d failure(s): %w", len(failures), errors.Join(failures...))
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
		if err := validateSyncEntry(ns, r); err != nil {
			return nil, err
		}
		entries[r.ID] = r
	}

	return entries, nil
}
