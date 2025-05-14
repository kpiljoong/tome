package journal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/paths"
)

var NamespacesCmd = &cobra.Command{
	Use:     "namespaces",
	Aliases: []string{"ns"},
	Short:   "List all namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		var namespaces []string

		from, _ := cmd.Flags().GetString(cliutil.FlagFrom)
		// Remote
		if from != "" {
			logx.Info("🔗 Resolving remote: %s", from)
			remote, err := cliutil.ResolveRemote(from, "")
			if err != nil {
				log.Fatalf("❌ Failed to resolve remote: %v", err)
			}
			namespaces, err = remote.ListNamespaces()
			if err != nil {
				log.Fatalf("❌ Failed to list namespaces from remote: %v", err)
			}
		} else {
			entries, err := os.ReadDir(paths.JournalsDir())
			if err != nil {
				log.Fatalf("❌ Could not read journals directory: %v", err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					namespaces = append(namespaces, entry.Name())
				}
			}
		}

		filter, _ := cmd.Flags().GetString("filter")
		if filter != "" {
			filtered := make([]string, 0, len(namespaces))
			for _, ns := range namespaces {
				if strings.Contains(strings.ToLower(ns), strings.ToLower(filter)) {
					filtered = append(filtered, ns)
				}
			}
			namespaces = filtered
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			if err := cliutil.PrintPrettyJSON(namespaces); err != nil {
				log.Fatalf("❌ Failed to encode JSON: %v", err)
			}
			return
		}

		quiet, _ := cmd.Flags().GetBool("quiet")
		if !quiet {
			logx.Info("📚 %d namespace(s) found:", len(namespaces))
		}

		for _, ns := range namespaces {
			fmt.Println("   📁", ns)
		}
	},
}

func init() {
	cliutil.AttachRemoteFlag(NamespacesCmd, cliutil.FlagFrom)
	cliutil.AttachQuietFlag(NamespacesCmd)
	cliutil.AttachJSONFlag(NamespacesCmd)
	NamespacesCmd.Flags().String("filter", "", "Substring to filter namespace names")
}
