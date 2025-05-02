package sync

import (
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/backend/s3"
	"github.com/kpiljoong/tome/internal/core"
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

		switch {
		case strings.HasPrefix(from, "s3://"):
			parts := strings.SplitN(strings.TrimPrefix(from, "s3://"), "/", 2)
			bucket := parts[0]
			prefix := ""
			if len(parts) > 1 {
				prefix = parts[1]
			}
			backend, err = s3.NewS3Backend(bucket, prefix)
			if err != nil {
				log.Fatalf("Failed to init S3 backend: %v", err)
			}
		default:
			log.Fatalf("Unsupported backend: %s", from)
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
