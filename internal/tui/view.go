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
	b.WriteString("\x1b[2J\x1b[H") // Clear + move cursor to top

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

		total := len(m.entries)
		visible := terminalHeight() - 5
		start := 0
		if m.cursor >= visible {
			start = m.cursor - visible + 1
		}
		end := start + visible
		if end > total {
			end = total
		}

		for i := start; i < end; i++ {
			cursor := "  "
			if i == m.cursor {
				cursor = "👉"
			}
			e := m.entries[i]
			b.WriteString(fmt.Sprintf("%s [%s] %-20s ID: %s\n",
				cursor,
				e.Timestamp.Format("2006-01-02 15:04"),
				e.Filename,
				e.ID[:10]))
		}

		b.WriteString("\n[↑↓] Move  [Enter] Select  [q] Quit")

	// file preview
	case stateFilePreview:
		b.WriteString(fmt.Sprintf("📄 Preview: %s\n\n", m.currentEntry.Filename))

		lines := strings.Split(m.preview, "\n")
		termHeight := terminalHeight()
		contentHeight := termHeight - 5

		if m.scrollOffset > len(lines)-1 {
			m.scrollOffset = len(lines) - 1
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}

		start := m.scrollOffset
		end := start + contentHeight
		if end > len(lines) {
			end = len(lines)
		}

		displayLines := lines[start:end]
		for _, line := range displayLines {
			if len(line) > 80 {
				line = line[:80]
			}
			b.WriteString(fmt.Sprintf("%-80s\n", line)) // enforce 80 width
		}

		// pad remaining
		for i := 0; i < contentHeight-len(displayLines); i++ {
			b.WriteString(strings.Repeat(" ", 80) + "\n")
		}

		b.WriteString("\n[↑↓] Scroll  [esc] Back  [q] Quit")
	}

	for strings.Count(b.String(), "\n") < terminalHeight() {
		b.WriteString("\n")
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
	return h - 4
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
