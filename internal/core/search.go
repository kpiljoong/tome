package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

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

	return entries, nil
}

func Search(namespace string, query string) ([]*model.JournalEntry, error) {
	baseDir := filepath.Join(os.Getenv("HOME"), ".tome")
	journalDir := filepath.Join(baseDir, "journals", namespace)

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

	return results, nil
}
