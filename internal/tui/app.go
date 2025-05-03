package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Start(initialNS string) error {
	m := initialModel(initialNS)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
