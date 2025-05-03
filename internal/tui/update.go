package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
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
					if m.cursor < m.viewport.YOffset {
						m.viewport.ScrollUp(1)
					}
					m.updateJournalViewport()
				}

			case stateFilePreview:
				m.viewport.ScrollUp(1)
			}

		case "down", "j":
			switch m.state {
			case stateNamespaceList:
				if m.cursor < len(m.namespaces)-1 {
					m.cursor++
					m.viewport.ScrollDown(1)
				}

			case stateJournalList:
				if m.cursor < len(m.entries)-1 {
					m.cursor++
					if m.cursor >= m.viewport.YOffset+m.viewport.Height {
						m.viewport.ScrollDown(1)
					}
					m.updateJournalViewport()
				}

			case stateFilePreview:
				m.viewport.ScrollDown(1)
			}

		case "esc", "backspace":
			switch m.state {
			case stateJournalList:
				m.state = stateNamespaceList
				m.cursor = 0

			case stateFilePreview:
				m.state = stateJournalList
				m.cursor = 0
				m.updateJournalViewport()
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

					m.updateJournalViewport()
					m.viewport.GotoTop()
				}

			case stateJournalList:
				if m.cursor < len(m.entries) {
					selected := m.entries[m.cursor]
					m.preview = loadFilePreview(selected)
					m.currentEntry = selected
					m.viewport.SetContent(m.preview)
					m.viewport.GotoTop()
					m.state = stateFilePreview
					m.cursor = 0
				}
			}
			return m, tea.ClearScreen
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			vp := viewport.New(msg.Width, msg.Height-2-4)
			vp.SetContent(m.preview)
			m.viewport = vp
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2 - 4
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

func (m *model) updateJournalViewport() {
	var b strings.Builder
	for i, e := range m.entries {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %-20s ID: %s\n",
			cursor,
			e.Timestamp.Format("2006-01-02 15:04"),
			e.Filename,
			e.ID[:10],
		))
	}
	m.viewport.SetContent(b.String())

	// if m.cursor < m.viewport.YOffset {
	// 	m.viewport.YOffset = m.cursor
	// } else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
	// 	m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	// }
}
