package journal

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
)

var from string

var SearchCmd = &cobra.Command{
	Use:   "search [namespace] [query]",
	Short: "Search journel entries by filename (fuzzy)",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var namespace, query string
		if len(args) >= 1 {
			namespace = args[0]
		}
		if len(args) == 2 {
			query = args[1]
		}

		from, _ := cmd.Flags().GetString(cliutil.FlagFrom)
		jsonOut, _ := cmd.Flags().GetBool(cliutil.FlagJSON)
		quiet, _ := cmd.Flags().GetBool(cliutil.FlagQuiet)
		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)

		var backend backend.RemoteBackend
		var allEntries []*model.JournalEntry
		var err error

		// Remote
		if from != "" {
			backend, err = cliutil.ResolveRemote(from, "")
			if err != nil {
				log.Fatalf("❌ Failed to resolve remote: %v", err)
			}
			namespaces := []string{namespace}
			if namespace == "" {
				namespaces, err = backend.ListNamespaces()
				if err != nil {
					log.Fatalf("❌ Failed to list namespaces from remote: %v", err)
				}
			}

			for _, ns := range namespaces {
				entries, err := backend.ListJournal(ns, query)
				if err == nil {
					allEntries = append(allEntries, entries...)
				}
				allEntries = append(allEntries, entries...)
			}
		} else {
			if namespace != "" {
				allEntries, _ = core.Search(namespace, query)
			} else {
				allEntries, _ = core.SearchAll(query)
			}
		}

		if len(allEntries) == 0 {
			if !quiet {
				logx.Warn("🚫 No entries found for query: %s", query)
			}
			return
		}

		if interactive {
			selected, err := cliutil.PickEntry(allEntries)
			if err != nil {
				log.Fatalf("❌ Failed to select entry: %v", err)
				return
			}
			allEntries = []*model.JournalEntry{selected}
		}

		if jsonOut {
			if err := cliutil.PrintPrettyJSON(allEntries); err != nil {
				log.Fatalf("❌ Failed to encode JSON: %v", err)
			}
			return
		}

		if query == "" && len(allEntries) > 0 && !interactive && !quiet {
			logx.Info("🕓 Latest entry:")
			fmt.Println(cliutil.FormatEntry(allEntries[0]))
			return
		}

		if !quiet {
			logx.Info("🔍 Found %d matching entries:", len(allEntries))
		}
		for _, e := range allEntries {
			fmt.Println(cliutil.FormatEntry(e))
		}
	},
}

func init() {
	cliutil.AttachJSONFlag(SearchCmd)
	cliutil.AttachQuietFlag(SearchCmd)
	cliutil.AttachInteractiveFlag(SearchCmd)
	cliutil.AttachRemoteFlag(SearchCmd, cliutil.FlagFrom)
	// SearchCmd.Flags().StringVar(&from, "from", "", "Optional remote backend to search from")
}
