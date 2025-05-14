package journal

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
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
	remote, err := cliutil.ResolveRemote(from, "")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote: %v", err)
	}

	entries, err := remote.ListJournal(namespace, query)
	if err != nil || len(entries) == 0 {
		return nil, fmt.Errorf("no matching journal entries found: %w", err)
	}

	var selected *model.JournalEntry
	if len(entries) == 1 {
		selected = entries[0]
	} else if interactive {
		selected, err = cliutil.PickEntry(entries)
		if err != nil {
			return nil, fmt.Errorf("entry selection failed: %w", err)
		}
	} else {
		logx.Warn("📘 🔍 %d matches found:", len(entries))
		for _, e := range entries {
			logx.Info("  - [%s] %-20s ID: %s", e.Timestamp.Format("2006-01-02 15:04"), e.Filename, e.ID[:8])
		}
		logx.Hint("Use '--interactive' to select one")
		return nil, fmt.Errorf("ambiguous result – refine your query")
	}

	return remote.GetBlobByHash(selected.BlobHash)
}

func init() {
	cliutil.AttachOutputFlag(GetCmd, "")
	cliutil.AttachJSONFlag(GetCmd)
	cliutil.AttachQuietFlag(GetCmd)
	cliutil.AttachRemoteFlag(GetCmd, cliutil.FlagFrom)
	cliutil.AttachInteractiveFlag(GetCmd)
}
