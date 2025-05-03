package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	coreModel "github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

type state int

const (
	stateLoading state = iota
	stateNamespaceList
	stateJournalList
	stateEntryDetail
	stateFilePreview
)

type model struct {
	state        state
	cursor       int
	scrollOffset int
	namespaces   []string
	entries      []*coreModel.JournalEntry
	currentNS    string
	currentEntry *coreModel.JournalEntry
	preview      string
}

func initialModel() model {
	journalDir := paths.JournalsDir()
	entries, err := os.ReadDir(journalDir)
	if err != nil {
		return model{state: stateNamespaceList, namespaces: []string{"(no namespaces found)"}, cursor: 0}
	}

	var namespaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			namespaces = append(namespaces, entry.Name())
		}
	}

	if len(namespaces) == 0 {
		namespaces = append(namespaces, "(empty)")
	}

	return model{
		state:      stateNamespaceList,
		namespaces: namespaces,
		cursor:     0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) listLen() int {
	switch m.state {
	case stateNamespaceList:
		return len(m.namespaces)

	case stateJournalList:
		return len(m.entries)

	default:
		return 0
	}
}
