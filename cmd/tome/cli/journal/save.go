package journal

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/util"
)

var SaveCmd = &cobra.Command{
	Use:   "save [namespace] [path]",
	Short: "Save a file or directory into the given namespace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		path := args[1]

		smart, _ := cmd.Flags().GetBool(cliutil.FlagSmart)
		excludePatterns, _ := cmd.Flags().GetStringArray(cliutil.FlagExclude)

		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("❌ Cannot access path: %v", err)
		}

		if info.IsDir() {
			entries, err := core.SaveDirWithExclude(namespace, path, smart, excludePatterns)
			if err != nil {
				log.Fatalf("❌ Failed to save directory: %v", err)
			}
			logx.Success("📦 Saved %d file(s) from directory: %s", len(entries), path)
			return
		}

		if util.ShouldExclude(path, excludePatterns) {
			logx.Info("🚫 Skipped excluded file: %s", path)
			return
		}

		entry, err := core.Save(namespace, path, smart)
		if err != nil {
			if strings.Contains(err.Error(), "already saved") {
				logx.Info("📎 Skipped (already saved): %s", path)
				return
			}
			log.Fatalf("❌ Error saving file: %v", err)
		}
		logx.Success("📘 Saved file: %s", path)
		logx.Info("📎 Namespace: %s", entry.Namespace)
		logx.Info("🧾 BlobHash:  %s", entry.BlobHash)
		logx.Info("🕓 Time:      %s", entry.Timestamp.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	cliutil.AttachSmartFlag(SaveCmd)
	cliutil.AttachExcludeFlag(SaveCmd)
}
