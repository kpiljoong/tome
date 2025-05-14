package journal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

var LatestCmd = &cobra.Command{
	Use:   "latest [namespace] [filename]",
	Short: "Get the latest file in a namespace",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ns := args[0]
		var filename string
		if len(args) == 2 {
			filename = strings.ToLower(args[1])
		}

		dir := paths.NamespaceDir(ns)
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatalf("Failed to read namespace dir: %v", err)
		}

		var matches []*model.JournalEntry
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				continue
			}
			var e model.JournalEntry
			if err := json.Unmarshal(data, &e); err != nil {
				continue
			}
			if filename == "" || strings.EqualFold(e.Filename, filename) {
				matches = append(matches, &e)
			}
		}

		if len(matches) == 0 {
			fmt.Println("No matching entries found.")
			return
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Timestamp.After(matches[j].Timestamp)
		})

		var selected *model.JournalEntry

		interactive, _ := cmd.Flags().GetBool(cliutil.FlagInteractive)
		if interactive {
			entry, err := cliutil.PickEntry(matches)
			if err != nil {
				log.Fatalf("%v", err)
			}
			selected = entry
		} else {
			selected = matches[0]
		}

		logx.Info("🕓 Latest entry: %s (%s)", selected.Filename, selected.Timestamp.Format("2006-01-02 15:04:05"))
		logx.Info("  ID:        %s", selected.ID)
		logx.Info("  BlobHash:  %s", selected.BlobHash)

		outputPath, _ := cmd.Flags().GetString(cliutil.FlagOutput)
		if outputPath != "" {
			data, err := os.ReadFile(paths.BlobPath(selected.BlobHash))
			if err != nil {
				log.Fatalf("Failed to read blob: %v", err)
			}

			if err := cliutil.WriteOutput(outputPath, data, false, false); err != nil {
				log.Fatalf("Failed to write output file: %v", err)
			}
			// fmt.Printf("Restored blob to %s\n", outputPath)
		}
	},
}

func init() {
	cliutil.AttachOutputFlag(LatestCmd, "")
	cliutil.AttachJSONFlag(LatestCmd)
	cliutil.AttachQuietFlag(LatestCmd)
	cliutil.AttachInteractiveFlag(LatestCmd)
}
