package tui

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

const maxVisibleLines = 30

func (m model) View() string {
	var b strings.Builder

	switch m.state {
	case stateNamespaceList:
		b.WriteString("📁 Namespaces\n\n")
		for i, ns := range m.namespaces {
			cursor := "  "
			if m.cursor == i {
				cursor = "👉"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, ns))
		}
		b.WriteString("\n[↑↓] Move  [Enter] Select  [q] Quit")

	case stateJournalList:
		b.WriteString(fmt.Sprintf("📓 Journals in %s\n\n", m.currentNS))
		b.WriteString(m.viewport.View())
		b.WriteString("\n[↑↓] Move  [Enter] Select  [Esc] Back  [q] Quit")

	// file preview
	case stateFilePreview:
		b.WriteString(fmt.Sprintf("📄 Preview: %s\n\n", m.currentEntry.Filename))

		b.WriteString(m.viewport.View())
		b.WriteString("\n[↑↓] Scroll  [esc] Back  [q] Quit")
	}

	return b.String()
}

func terminalHeight() int {
	h, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || h < 5 {
		return maxVisibleLines
	}
	if h > maxVisibleLines {
		return maxVisibleLines
	}
	return h
}

func visibleLines(s string, termWidth int) int {
	lines := 0
	for _, line := range strings.Split(s, "\n") {
		runeCount := utf8.RuneCountInString(line)
		if runeCount == 0 {
			lines++
		} else {
			lines += (runeCount / termWidth) + 1
		}
	}
	return lines
}

func terminalWidth() int {
	_, w, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 20 {
		return 80 // fallback default
	}
	return w
}
