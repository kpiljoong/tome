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
		ns, _ := cmd.Flags().GetString("namespace")
		if err := tui.Start(ns); err != nil {
			log.Fatalf("❌ TUI failed: %v", err)
		}
	},
}

func init() {
	TuiCmd.Flags().String("namespace", "", "Start directly in a given namespace")
}
