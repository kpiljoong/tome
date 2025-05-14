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

var ListCmd = &cobra.Command{
	Use:     "list [namespace]",
	Aliases: []string{"ls"},
	Short:   "List all files in a namespace",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := args[0]
		dir := paths.NamespaceDir(ns)

		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatalf("Failed to read namespace dir: %v", err)
		}

		var entries []*model.JournalEntry
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				continue
			}
			var e model.JournalEntry
			if err := json.Unmarshal(data, &e); err == nil {
				entries = append(entries, &e)
			}
		}

		if len(entries) == 0 {
			logx.Warn("🚫 No entries found in namespace: %s", ns)
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.After(entries[j].Timestamp)
		})

		jsonOut, _ := cmd.Flags().GetBool(cliutil.FlagJSON)
		if jsonOut {
			if err := cliutil.PrintPrettyJSON(entries); err != nil {
				logx.Error("Failed to encode JSON: %v", err)
			}
		} else {
			for _, e := range entries {
				fmt.Printf("%s  %-20s  %s\n",
					e.Timestamp.Format("2006-01-02 15:04"),
					e.Filename,
					e.ID)
			}
		}
	},
}

func init() {
	cliutil.AttachJSONFlag(ListCmd)
}
