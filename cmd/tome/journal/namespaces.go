package journal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var NamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "List all namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Could not determine home dir: %v", err)
		}
		base := filepath.Join(home, ".tome", "journals")

		entries, err := os.ReadDir(base)
		if err != nil {
			log.Fatalf("Could not read base directory: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				fmt.Println(entry.Name())
			}
		}
	},
}
