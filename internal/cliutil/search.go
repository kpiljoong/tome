package cliutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

func LocalSearch(namespace, query string) ([]*model.JournalEntry, error) {
	nsDir := paths.NamespaceDir(namespace)
	files, err := os.ReadDir(nsDir)
	if err != nil {
		return nil, err
	}

	var matches []*model.JournalEntry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}

		path := filepath.Join(nsDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if query == "" ||
			strings.EqualFold(entry.Filename, query) ||
			strings.EqualFold(entry.FullPath, query) {
			matches = append(matches, &entry)
		}
	}
	return matches, nil
}
