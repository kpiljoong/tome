package cli

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/util"
)

var ShareCmd = &cobra.Command{
	Use:   "share [namespace] [filename]",
	Short: "Generate a temporary public link for a file in remote store",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace := args[0]
		filename := args[1]

		from, _ := cmd.Flags().GetString(cliutil.FlagFrom)
		if from == "" {
			log.Fatalf("❌ --from flag is required (e.g., s3://bucket/prefix)")
		}

		expiresStr, _ := cmd.Flags().GetString("expires")
		duration, err := time.ParseDuration(expiresStr)
		if err != nil {
			log.Fatalf("❌ Invalid duration format: %v", err)
		}

		// if !strings.HasPrefix(from, "s3://") {
		// 	log.Fatalf("❌ Currently only S3 backends are supported for sharing")
		// }

		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)
		shorten, _ := cmd.Flags().GetBool(cliutil.FlagShorten)

		remote, err := cliutil.ResolveRemote(from, "")
		if err != nil {
			log.Fatalf("❌ Failed to resolve remote backend: %v", err)
		}

		entries, err := remote.ListJournal(namespace, filename)
		if err != nil || len(entries) == 0 {
			log.Fatalf("❌ Failed to list entries in namespace %s: %v", namespace, err)
		}

		var selected *model.JournalEntry
		// if len(entries) > 1 {
		// 	logx.Info("🔍 %d matches found for %q in namespace [%s]:", len(entries), filename, namespace)
		// 	for _, e := range entries {
		// 		fmt.Printf("  - [%s] %-20s  ID: %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.Filename, e.ID[:8])
		// 	}
		// 	logx.Hint("Use '--interactive' to pick one.")
		// 	log.Fatalf("❌ Multiple results. Please refine your query.")
		// }
		//
		// entry := entries[0]
		// key := remote.BlobKey(entry.BlobHash)

		if len(entries) == 1 {
			selected = entries[0]
		} else if interactive {
			selected, err = cliutil.PickEntry(entries)
			if err != nil {
				log.Fatalf("❌ Failed to select entry: %v", err)
			}
		} else {
			logx.Info("🔍 %d matches found for %q in namespace [%s]:", len(entries), filename, namespace)
			for _, e := range entries {
				fmt.Printf("  - [%s] %-20s  ID: %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.Filename, e.ID[:8])
			}
			logx.Hint("Use '--interactive' to pick one.")
			log.Fatalf("❌ Multiple results. Please refine your query.")
		}

		key := remote.BlobKey(selected.BlobHash)
		url, err := remote.GeneratePresignedURL(key, duration)
		if err != nil {
			log.Fatalf("❌ Failed to generate presigned URL: %v", err)
		}

		if shorten {
			shortURL, err := util.ShortenURL(url)
			if err != nil {
				logx.Error("Failed to shorten URL: %v", err)
			} else {
				logx.Success("🔗 Shortened URL: %s", shortURL)
				return
			}
		}

		logx.Success("🔗 Shareable link (valid %s):\n%s", expiresStr, url)
	},
}

func init() {
	ShareCmd.Flags().String("expires", "10m", "Duration for link validity (e.g., 10m, 1h)")
	cliutil.AttachRemoteFlag(ShareCmd, cliutil.FlagFrom)
	cliutil.AttachShortenFlag(ShareCmd)
	cliutil.AttachInteractiveFlag(ShareCmd)
}
