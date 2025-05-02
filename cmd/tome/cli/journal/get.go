package journal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/backend/s3"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/cliutil"
	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
)

var GetCmd = &cobra.Command{
	Use:   "get [namespace] [query or full path]",
	Short: "Get a file from the given namespace",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace, query := args[0], args[1]

		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)
		outputPath, _ := cmd.Flags().GetString(cliutil.FlagOutput)
		from, _ := cmd.Flags().GetString(cliutil.FlagFrom)
		quiet, _ := cmd.Flags().GetBool(cliutil.FlagQuiet)
		jsonOut, _ := cmd.Flags().GetBool(cliutil.FlagJSON)

		var data []byte
		var err error

		if from != "" {
			data, err = getFromRemote(namespace, query, from, interactive)
		} else {
			data, err = getFromLocal(namespace, query, interactive)
		}

		if err != nil {
			log.Fatalf("Get failed: %v\n", err)
		}

		if err := cliutil.WriteOutput(outputPath, data, quiet, jsonOut); err != nil {
			log.Fatalf(" %v", err)
		}
	},
}

func getFromLocal(namespace, query string, interactive bool) ([]byte, error) {
	entries, err := core.SearchLocal(namespace, query)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no matching entries found")
	}

	var selected *model.JournalEntry
	if len(entries) == 1 {
		selected = entries[0]
	} else if interactive {
		selected, err = cliutil.PickEntry(entries)
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Multiple matches found:")
		for _, e := range entries {
			fmt.Printf("  - [%s] %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.FullPath)
		}
		fmt.Println("💡 Tip: use '--interactive' to select one")
		return nil, fmt.Errorf("ambiguous result - refine your query")
	}

	return os.ReadFile(paths.BlobPath(selected.BlobHash))
}

func getFromRemote(namespace, query, from string, interactive bool) ([]byte, error) {
	if !strings.HasPrefix(from, "s3://") {
		return nil, fmt.Errorf("unsupported backend: %s", from)
	}

	parts := strings.SplitN(strings.TrimPrefix(from, "s3://"), "/", 2)
	bucket := parts[0]
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[1]
	}

	backend, err := s3.NewS3Backend(bucket, prefix)
	if err != nil {
		log.Fatalf("S3 backend init failed: %v", err)
	}

	entries, err := backend.ListJournal(namespace, query)
	if err != nil || len(entries) == 0 {
		log.Fatalf("Failed to list journal entries: %v", err)
	}

	var selected *model.JournalEntry
	if len(entries) == 1 {
		selected = entries[0]
	} else if interactive {
		selected, err = cliutil.PickEntry(entries)
		if err != nil {
			return nil, err
		}
	} else {
		logx.Warn("Multiple matches found:")
		for _, e := range entries {
			fmt.Printf("   🧾  [%s] %s\n", e.Timestamp.Format("2006-01-02 15:04"), e.FullPath)
		}
		logx.Hint("Use --interactive to pick one")
		log.Fatalf("❌ Ambiguous result — refine your query")
	}

	return core.GetBlobByHash(selected.BlobHash)
}

func init() {
	cliutil.AttachOutputFlag(GetCmd, "")
	cliutil.AttachJSONFlag(GetCmd)
	cliutil.AttachQuietFlag(GetCmd)
	cliutil.AttachRemoteFlag(GetCmd, cliutil.FlagFrom)
	cliutil.AttachInteractiveFlag(GetCmd)
}
