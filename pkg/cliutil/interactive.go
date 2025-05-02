package cliutil

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/kpiljoong/tome/pkg/model"
)

// PickEntry lets the user choose an entry from a list using terminal input.
func PickEntry(entries []*model.JournalEntry) (*model.JournalEntry, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries available")
	}

	fmt.Println("Select an entry:")
	for i, e := range entries {
		fmt.Printf("%2d. %s  %-20s  %s\n",
			i+1,
			e.Timestamp.Format("2006-01-02 15:04"),
			e.Filename,
			e.ID[:8],
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
