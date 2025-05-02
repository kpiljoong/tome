package journal

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend/s3"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/model"
)

var from string

var SearchCmd = &cobra.Command{
	Use:   "search [namespace] [query]",
	Short: "Search for files in the given namespace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace, query := args[0], args[1]

		var results []*model.JournalEntry
		var err error

		if from != "" {
			switch {
			case strings.HasPrefix(from, "s3://"):
				parts := strings.SplitN(strings.TrimPrefix(from, "s3://"), "/", 2)
				bucket := parts[0]
				prefix := ""
				if len(parts) > 1 {
					prefix = parts[1]
				}
				backend, err := s3.NewS3Backend(bucket, prefix)
				if err != nil {
					log.Fatalf("S3 backend init failed: %v", err)
				}
				results, err = backend.ListJournal(namespace, query)
			default:
				log.Fatalf("Unknown backend target: %s", from)
			}
		} else {
			results, err = core.Search(namespace, query)
		}

		if err != nil {
			fmt.Printf("Error searching for files: %v\n", err)
			return
		}

		fmt.Printf("Found %d entries\n", len(results))

		for _, entry := range results {
			fmt.Printf("[%s] %s\n", entry.Timestamp.Format("2006-01-02 15:04"), entry.FullPath)
		}
	},
}

func init() {
	SearchCmd.Flags().StringVar(&from, "from", "", "Optional remote backend to search from")
}
