package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
	"github.com/kpiljoong/tome/pkg/util"
)

func SearchAll(query string) ([]*model.JournalEntry, error) {
	root := paths.JournalsDir()
	namespaces, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read journals root: %w", err)
	}

	var all []*model.JournalEntry
	for _, ns := range namespaces {
		if !ns.IsDir() {
			continue
		}

		nsName := ns.Name()
		entries, err := Search(nsName, query)
		if err == nil {
			all = append(all, entries...)
		}
	}

	util.SortEntriesByTimestampDesc(all)

	return all, nil
}

func SearchLocal(namespace, query string) ([]*model.JournalEntry, error) {
	dir := paths.NamespaceDir(namespace)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read namespace dir: %w", err)
	}

	var entries []*model.JournalEntry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err == nil {
			if strings.EqualFold(entry.Filename, query) || strings.EqualFold(entry.FullPath, query) {
				entries = append(entries, &entry)
			}
		}
	}

	util.SortEntriesByTimestampDesc(entries)

	return entries, nil
}

func Search(namespace string, query string) ([]*model.JournalEntry, error) {
	// baseDir := filepath.Join(os.Getenv("HOME"), ".tome")
	// journalDir := filepath.Join(baseDir, "journals", namespace)
	journalDir := paths.NamespaceDir(namespace)

	var results []*model.JournalEntry

	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read journal dir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		entryPath := filepath.Join(journalDir, file.Name())
		data, err := os.ReadFile(entryPath)
		if err != nil {
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(entry.Filename), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(entry.FullPath), strings.ToLower(query)) {
			results = append(results, &entry)
		}
	}

	util.SortEntriesByTimestampDesc(results)

	return results, nil
}
