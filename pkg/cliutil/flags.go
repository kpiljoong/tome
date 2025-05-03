package cliutil

import "github.com/spf13/cobra"

const (
	FlagOutput      = "output"
	FlagInteractive = "interactive"
	FlagFrom        = "from"
	FlagTo          = "to"
	FlagMode        = "mode"
	FlagJSON        = "json"
	FlagQuiet       = "quiet"
	FlagLimit       = "limit"
	FlagSmart       = "smart"
	FlagShorten     = "shorten"
	FlagExclude     = "exclude"
	FlagAll         = "all"
)

func AttachOutputFlag(cmd *cobra.Command, defaultPath string, help ...string) *string {
	helpText := "Write blob content to this path"
	if len(help) > 0 {
		helpText = help[0]
	}
	return cmd.Flags().StringP(FlagOutput, "o", defaultPath, helpText)
}

func AttachJSONFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagJSON, false, "Output in JSON format")
}

func AttachQuietFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagQuiet, false, "Suppress output")
}

func AttachLimitFlag(cmd *cobra.Command) *string {
	return cmd.Flags().String(FlagLimit, "", "Limit number of entries shown")
}

func AttachRemoteFlag(cmd *cobra.Command, name string) *string {
	return cmd.Flags().String(name, "", "Remote backend path (e.g. s3://bucket/prefix)")
}

func AttachModeFlag(cmd *cobra.Command) *string {
	return cmd.Flags().String(FlagMode, "push", "Sync mode: push, pull, sync")
}

func AttachInteractiveFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagInteractive, false, "Enable interactive mode")
}

func AttachSmartFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagSmart, false, "Enable smart mode")
}

func AttachShortenFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagShorten, false, "Enable shorten mode")
}

func AttachExcludeFlag(cmd *cobra.Command) *[]string {
	return cmd.Flags().StringArray(FlagExclude, nil, "Exclude entries matching this pattern")
}

func AttachAllFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().Bool(FlagAll, false, "Include all entries")
}
