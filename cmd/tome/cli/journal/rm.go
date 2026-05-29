package journal

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

var errAmbiguousRmMatch = errors.New("ambiguous matches")

var RmCmd = &cobra.Command{
	Use:   "rm [namespace] [query]",
	Short: "Remove matching journal entries from local store",
	Args:  validateRmArgs,
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		query := ""
		if len(args) >= 2 {
			query = args[1]
		}

		allFlag, _ := cmd.Flags().GetBool(cliutil.FlagAll)
		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)

		entries, err := cliutil.LocalSearch(namespace, query)
		if err != nil {
			log.Fatalf("❌ Failed to search journal entries: %v", err)
		}
		if len(entries) == 0 {
			log.Fatalf("❌ No matching entries found for '%s' in namespace '%s'", query, namespace)
		}

		toDelete, aborted, err := resolveRmTargets(namespace, query, entries, allFlag, interactive, func() bool {
			return confirmDeleteAll(namespace)
		}, cliutil.PickEntry)
		if err != nil {
			if errors.Is(err, errAmbiguousRmMatch) {
				logx.Info("🔍 Multiple matches found for %q in namespace [%s]:", query, namespace)
				for _, e := range entries {
					logx.Info("  - [%s] %-20s  ID: %s", e.Timestamp.Format("2006-01-02 15:04"), e.Filename, e.ID[:8])
				}
				logx.Hint("Use '--all' to delete all, or '--interactive' to pick one.")
				log.Fatalf("❌ Ambiguous match — refine query or use --all/--interactive")
			}
			log.Fatalf("❌ Failed to select entries: %v", err)
		}
		if aborted {
			logx.Warn("Aborted by user.")
			return
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

func validateRmArgs(cmd *cobra.Command, args []string) error {
	allFlag, _ := cmd.Flags().GetBool(cliutil.FlagAll)
	if allFlag {
		if len(args) < 1 || len(args) > 2 {
			return fmt.Errorf("requires [namespace] and optional [query] when --all is used")
		}
		return nil
	}
	return cobra.ExactArgs(2)(cmd, args)
}

func resolveRmTargets(
	namespace string,
	query string,
	entries []*model.JournalEntry,
	allFlag bool,
	interactive bool,
	confirmAll func() bool,
	pickEntry func([]*model.JournalEntry) (*model.JournalEntry, error),
) ([]*model.JournalEntry, bool, error) {
	switch {
	case allFlag:
		if query == "" && !confirmAll() {
			return nil, true, nil
		}
		return entries, false, nil
	case len(entries) == 1:
		return entries, false, nil
	case interactive:
		selected, err := pickEntry(entries)
		if err != nil {
			return nil, false, err
		}
		return []*model.JournalEntry{selected}, false, nil
	default:
		return nil, false, errAmbiguousRmMatch
	}
}

func confirmDeleteAll(namespace string) bool {
	fmt.Printf("⚠️  Are you sure you want to delete ALL entries in namespace [%s]? (y/N): ", namespace)
	var input string
	fmt.Scanln(&input)
	return strings.ToLower(input) == "y"
}
