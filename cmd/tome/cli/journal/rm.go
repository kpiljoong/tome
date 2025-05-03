package journal

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

var RmCmd = &cobra.Command{
	Use:   "rm [namespace] [query]",
	Short: "Remove matching journal entries from local store",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		query := args[1]

		allFlag, _ := cmd.Flags().GetBool(cliutil.FlagAll)
		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)

		entries, err := cliutil.LocalSearch(namespace, query)
		if err != nil {
			log.Fatalf("❌ Failed to search journal entries: %v", err)
		}
		if len(entries) == 0 {
			log.Fatalf("❌ No matching entries found for '%s' in namespace '%s'", query, namespace)
		}

		var toDelete []*model.JournalEntry

		switch {
		case len(entries) == 1:
			toDelete = entries

		case allFlag:
			fmt.Printf("⚠️  Are you sure you want to delete ALL entries in namespace [%s]? (y/N): ", namespace)
			var input string
			fmt.Scanln(&input)
			if strings.ToLower(input) != "y" {
				logx.Warn("Aborted by user.")
				return
			}
			toDelete = entries

		case interactive:
			selected, err := cliutil.PickEntry(entries)
			if err != nil {
				log.Fatalf("❌ Failed to select entry: %v", err)
			}
			toDelete = []*model.JournalEntry{selected}

		default:
			logx.Info("🔍 Multiple matches found for %q in namespace [%s]:", query, namespace)
			for _, e := range entries {
				logx.Info("  - [%s] %-20s  ID: %s", e.Timestamp.Format("2006-01-02 15:04"), e.Filename, e.ID[:8])
			}
			logx.Hint("Use '--all' to delete all, or '--interactive' to pick one.")
			log.Fatalf("❌ Ambiguous match — refine query or use --all/--interactive")
		}

		for _, entry := range toDelete {
			path := paths.JournalEntryPath(entry)
			if err := cliutil.SafeDelete(path); err != nil {
				logx.Error("❌ Failed to delete entry %s: %v", path, err)
			} else {
				logx.Success("🗑️  Deleted %s", entry.Filename)
			}
		}
	},
}

func init() {
	cliutil.AttachInteractiveFlag(RmCmd)
	cliutil.AttachAllFlag(RmCmd)
}
