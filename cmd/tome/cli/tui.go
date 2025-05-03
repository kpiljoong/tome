package cli

import (
	"log"

	"github.com/kpiljoong/tome/internal/tui"
	"github.com/spf13/cobra"
)

var TuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the TUI interface",
	Run: func(cmd *cobra.Command, args []string) {
		if err := tui.Start(); err != nil {
			log.Fatalf("❌ TUI failed: %v", err)
		}
	},
}
