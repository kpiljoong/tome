package cliutil

import "github.com/spf13/cobra"

const (
	FlagOutput      = "output"
	FlagInteractive = "interactive"
	FlagFrom        = "from"
	FlagTo          = "to"
	FlagMode        = "mode"
)

func AttachOutputFlag(cmd *cobra.Command, defaultPath string, help ...string) *string {
	helpText := "Write blob content to this path"
	if len(help) > 0 {
		helpText = help[0]
	}
	return cmd.Flags().StringP(FlagOutput, "o", defaultPath, helpText)
}

func AttachJSONFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool("json", false, "Output in JSON format")
}

func AttachRemoteFlag(cmd *cobra.Command, name string) *string {
	return cmd.Flags().String(name, "", "Remote backend path (e.g. s3://bucket/prefix)")
}

func AttachModeFlag(cmd *cobra.Command) *string {
	return cmd.Flags().String("mode", "push", "Sync mode: push, pull, sync")
}

func AttachInteractiveFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagInteractive, false, "Enable interactive mode")
}
