package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Start() error {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
