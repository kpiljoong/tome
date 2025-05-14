package cliutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

// PickEntry lets the user choose an entry from a list using terminal input.
func PickEntry(entries []*model.JournalEntry) (*model.JournalEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries available")
	}

	fmt.Println("🧾 Select a journal entry:")
	for i, e := range entries {
		preview := generateSmartPreview(e)
		fmt.Printf(" %2d. [%-16s]  %-20s  ID: %.8s  → %s\n",
			i+1,
			e.Timestamp.Format("2006-01-02 15:04"),
			e.Filename,
			e.ID[:8],
			preview,
		)
	}

	fmt.Print("Enter number: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %v", err)
	}

	num, err := strconv.Atoi(line[:len(line)-1])
	if err != nil || num < 1 || num > len(entries) {
		return nil, fmt.Errorf("invalid entry number: %v", err)
	}
	return entries[num-1], nil
}

func generateSmartPreview(e *model.JournalEntry) string {
	blobPath := paths.BlobPath(e.BlobHash)
	data, err := os.ReadFile(blobPath)
	if err != nil {
		return "[error reading blob]"
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		var parts []string
		for k, v := range m {
			str := fmt.Sprintf("%s=%v", k, v)
			if len(str) > 30 {
				str = str[:30] + "..."
			}
			parts = append(parts, str)
		}
		sort.Strings(parts)
		if len(parts) > 3 {
			parts = parts[:3]
		}
		return strings.Join(parts, " ")
	}

	// fallback
	lines := strings.Split(string(data), "\n")
	var preview []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if len(l) > 50 {
			l = l[:50] + "..."
		}
		preview = append(preview, l)
		if len(preview) >= 2 {
			break
		}
	}
	if len(preview) == 0 {
		return "[empty file]"
	}
	return strings.Join(preview, " ")
}
