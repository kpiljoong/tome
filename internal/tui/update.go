package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kpiljoong/tome/pkg/logx"
	coreModel "github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			switch m.state {
			case stateNamespaceList, stateJournalList:
				if m.cursor > 0 {
					m.cursor--
				}

			case stateFilePreview:
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			}

		case "down", "j":
			switch m.state {
			case stateNamespaceList:
				if m.cursor < len(m.namespaces)-1 {
					m.cursor++
				}

			case stateJournalList:
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}

			case stateFilePreview:
				lines := strings.Split(m.preview, "\n")
				maxScroll := len(lines) - terminalHeight() + 5
				if m.scrollOffset < maxScroll {
					m.scrollOffset++
				}
			}

		case "esc", "backspace":
			switch m.state {
			case stateJournalList:
				m.state = stateNamespaceList
				m.cursor = 0

			case stateFilePreview:
				m.state = stateJournalList
				m.cursor = 0
			}
			return m, tea.ClearScreen

		case "enter":
			switch m.state {
			case stateNamespaceList:
				if len(m.namespaces) > 0 {
					selectedNS := m.namespaces[m.cursor]
					journalEntries := loadJournalEntries(selectedNS)
					m.state = stateJournalList
					m.currentNS = selectedNS
					m.entries = journalEntries
					m.cursor = 0
				}

			case stateJournalList:
				if m.cursor < len(m.entries) {
					selected := m.entries[m.cursor]
					m.preview = loadFilePreview(selected)
					m.currentEntry = selected
					m.state = stateFilePreview
					m.cursor = 0
				}
			}
			return m, tea.ClearScreen
		}
	}
	return m, nil
}

func loadJournalEntries(namespace string) []*coreModel.JournalEntry {
	nsDir := paths.NamespaceDir(namespace)
	files, err := os.ReadDir(nsDir)
	if err != nil {
		logx.Warn("Failed to read namespace directory: %v", err)
		return []*coreModel.JournalEntry{}
	}

	var entries []*coreModel.JournalEntry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(nsDir, f.Name()))
		if err != nil {
			continue
		}

		var entry coreModel.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		entries = append(entries, &entry)
	}

	return entries
}

func loadFilePreview(e *coreModel.JournalEntry) string {
	blobPath := paths.BlobPath(e.BlobHash)
	data, err := os.ReadFile(blobPath)
	if err != nil {
		return fmt.Sprintf("❌ Failed to read blob: %v", err)
	}
	return string(data)
}
