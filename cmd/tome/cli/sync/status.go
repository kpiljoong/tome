package sync

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/paths"
)

var (
	jsonOut bool
	from    string
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the local and remote entries",
	Long:  `Show the status of the local and remote entries. This command will show the status of the local and remote entries in the specified namespace. It will show if the entry is only in local, only in remote, or synced.`,
	Run: func(cmd *cobra.Command, args []string) {
		localPath := paths.TomeRoot()

		if from == "" {
			log.Fatalf("Must specify --from for remote comparison")
		}

		var backend backend.RemoteBackend
		var err error

		backend, err = cliutil.ResolveRemote(from, "")
		if err != nil {
			logx.Error("Failed to resolve remote: %v", err)
			log.Fatalf("Sync aborted")
		}

		if err := core.Status(localPath, backend, jsonOut); err != nil {
			log.Fatalf("Status failed: %v", err)
		}
	},
}

func init() {
	StatusCmd.Flags().StringVar(&from, "from", "", "Remote backend to compare against (e.g. s3://bucket/prefix)")
	StatusCmd.Flags().BoolVar(&jsonOut, "json", false, "Output in JSON format")
}
