package cliutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
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

		paths := filepath.Join(nsDir, f.Name())
		data, err := os.ReadFile(paths)
		if err != nil {
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if strings.Contains(entry.Filename, query) {
			matches = append(matches, &entry)
		}
	}
	return matches, nil
}
