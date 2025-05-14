package sync

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/config"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/paths"
)

var SyncCmd = &cobra.Command{
	Use:   "sync --to [s3://bucket/path | github://org/repo]",
	Short: "Sync journal to a remote static store",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var backend backend.RemoteBackend

		target, _ := cmd.Flags().GetString(cliutil.FlagTo)
		mode, _ := cmd.Flags().GetString(cliutil.FlagMode)
		cfg, _ := config.Load()

		backend, err = cliutil.ResolveRemote(target, cfg.DefaultRemote)
		if err != nil {
			logx.Error("Failed to resolve remote: %v", err)
			log.Fatalf("Sync aborted")
		}

		logx.Info("Mode: %s -> %s", mode, backend.Describe())

		switch mode {
		case "push":
			logx.Info("📤 Pushing local data to remote...")
			err = core.Sync(paths.TomeRoot(), backend)
		case "pull":
			logx.Info("📥 Pulling from remote to local...")
			err = core.Pull(paths.TomeRoot(), backend)
		case "sync":
			logx.Info("🔄 Bidirectional sync...")
			err = core.SyncBidirectional(paths.TomeRoot(), backend)
		default:
			logx.Error("Unknown sync mode: %s", mode)
			// log.Fatalf("Unknown sync mode: %s", mode)
		}

		if err != nil {
			logx.Error("Sync failed: %v", err)
			log.Fatalf("Sync aborted")
			// log.Fatalf("Sync failed: %v", err)
		}
		logx.Success("✅ Sync complete")
	},
}

func init() {
	cliutil.AttachRemoteFlag(SyncCmd, cliutil.FlagTo)
	cliutil.AttachModeFlag(SyncCmd)
}
