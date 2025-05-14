package journal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

var jsonOut bool

var ShowCmd = &cobra.Command{
	Use:   "show [namespace] [id]",
	Short: "Show metadata of a journal entry",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ns, id := args[0], args[1]
		path := paths.JournalPath(ns, id)

		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(entry)
		} else {
			fmt.Printf("ID:         %s\n", entry.ID)
			fmt.Printf("Namespace:  %s\n", entry.Namespace)
			fmt.Printf("Filename:   %s\n", entry.Filename)
			fmt.Printf("Full Path:  %s\n", entry.FullPath)
			fmt.Printf("Timestamp:  %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("Blob Hash:  %s\n", entry.BlobHash)
			fmt.Printf("Meta:\n")
			for k, v := range entry.Meta {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	},
}

func init() {
	ShowCmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
}
