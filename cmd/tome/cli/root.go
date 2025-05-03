package cli

import (
	"github.com/kpiljoong/tome/cmd/tome/cli/journal"
	"github.com/kpiljoong/tome/cmd/tome/cli/serve"
	"github.com/kpiljoong/tome/cmd/tome/cli/sync"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tome",
	Short: "Tome is a zero-runtime journal and blob store",
	Long:  `Tome lets you save, search, and retrieve files into a local or remote append-only store.`,
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.AddCommand(
		journal.GetCmd,
		journal.LatestCmd,
		journal.ListCmd,
		journal.LogCmd,
		journal.NamespacesCmd,
		journal.SaveCmd,
		journal.SearchCmd,
		journal.ShowCmd,
		journal.RmCmd,
		sync.StatusCmd,
		sync.SyncCmd,
		serve.ServeCmd,
		configCmd,
		ShareCmd,
		TuiCmd,
	)
}
