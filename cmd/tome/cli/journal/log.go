package journal

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/util"
)

var LogCmd = &cobra.Command{
	Use:   "log [namespace] [query]",
	Short: "Show recent saved entries (like git log)",
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
		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)
		quiet, _ := cmd.Flags().GetBool(cliutil.FlagQuiet)
		limitStr, _ := cmd.Flags().GetString(cliutil.FlagLimit)

		limit := 0
		if limitStr != "" {
			limit, _ = strconv.Atoi(limitStr)
		}

		var entries []*model.JournalEntry

		if from != "" {
			remote, err := cliutil.ResolveRemote(from, "")
			if err != nil {
				logx.Error("❌ Failed to resolve remote: %v", err)
			}
			nsList := []string{namespace}
			if namespace == "" {
				nsList, _ = remote.ListNamespaces()
			}
			for _, ns := range nsList {
				nsEntries, _ := remote.ListJournal(ns, query)
				entries = append(entries, nsEntries...)
			}
		} else {
			if namespace != "" {
				entries, _ = core.Search(namespace, query)
			} else {
				entries, _ = core.SearchAll(query)
			}
		}

		util.SortEntriesByTimestampDesc(entries)

		if limit > 0 && len(entries) > limit {
			entries = entries[:limit]
		}

		if len(entries) == 0 && !quiet {
			logx.Warn("🚫 No entries found in namespace: %s", namespace)
			return
		}

		if interactive {
			selected, err := cliutil.PickEntry(entries)
			if err != nil {
				logx.Error("❌ Interactive selection failed: %v", err)
				return
			}
			entries = []*model.JournalEntry{selected}
		}

		if jsonOut {
			cliutil.PrintPrettyJSON(entries)
			return
		}

		for _, e := range entries {
			fmt.Println(cliutil.FormatEntry(e))
		}
	},
}

func init() {
	cliutil.AttachRemoteFlag(LogCmd, cliutil.FlagFrom)
	cliutil.AttachJSONFlag(LogCmd)
	cliutil.AttachQuietFlag(LogCmd)
	cliutil.AttachInteractiveFlag(LogCmd)
	cliutil.AttachLimitFlag(LogCmd)
}
