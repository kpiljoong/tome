package main

import (
	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/cmd/tome/journal"
	"github.com/kpiljoong/tome/cmd/tome/serve"
	"github.com/kpiljoong/tome/cmd/tome/sync"
)

var rootCmd = &cobra.Command{
	Use:   "tome",
	Short: "Tome is a zero-runtime journal and blob store",
	Long:  `Tome lets you save, search, and retrieve files into a local or remote append-only store.`,
}

func init() {
	rootCmd.AddCommand(
		journal.GetCmd,
		journal.LatestCmd,
		journal.ListCmd,
		journal.NamespacesCmd,
		journal.SaveCmd,
		journal.SearchCmd,
		journal.ShowCmd,
		sync.StatusCmd,
		sync.SyncCmd,
		serve.ServeCmd,
		configCmd,
	)
}
