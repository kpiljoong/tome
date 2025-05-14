package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	coreModel "github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
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
	state      state
	namespaces []string
	cursor     int

	entries      []*coreModel.JournalEntry
	currentNS    string
	currentEntry *coreModel.JournalEntry

	preview  string
	viewport viewport.Model
	ready    bool
}

func initialModel(optionalNS ...string) model {
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

	m := model{
		state:      stateNamespaceList,
		namespaces: namespaces,
		cursor:     0,
	}

	if len(optionalNS) > 0 {
		ns := optionalNS[0]
		for _, n := range namespaces {
			if n == ns {
				m.state = stateJournalList
				m.currentNS = ns
				m.entries = loadJournalEntries(ns)
				break
			}
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
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
