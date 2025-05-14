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
	return attachStringFlag(cmd, FlagOutput, "o", defaultPath, help...)
}

func AttachJSONFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagJSON, "", false, "Output in JSON format")
}

func AttachQuietFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagQuiet, "", false, "Suppress output")
}

func AttachLimitFlag(cmd *cobra.Command) *string {
	return attachStringFlag(cmd, FlagLimit, "", "", "Limit number of entries shown")
}

func AttachRemoteFlag(cmd *cobra.Command, name string) *string {
	return attachStringFlag(cmd, name, "", "", "Remote backend path (e.g. s3://bucket/prefix)")
}

func AttachModeFlag(cmd *cobra.Command) *string {
	return attachStringFlag(cmd, FlagMode, "", "push", "Sync mode: push, pull, sync")
}

func AttachInteractiveFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagInteractive, "", false, "Enable interactive mode")
}

func AttachSmartFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagSmart, "", false, "Enable smart mode")
}

func AttachShortenFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagShorten, "", false, "Enable shorten mode")
}

func AttachExcludeFlag(cmd *cobra.Command) *[]string {
	return attachStringArrayFlag(cmd, FlagExclude, "e", nil, "Exclude entries matching this pattern")
}

func AttachAllFlag(cmd *cobra.Command) *bool {
	return attachBoolFlag(cmd, FlagAll, "a", false, "Include all entries")
}

// Helper function to attach a boolean flag
func attachBoolFlag(cmd *cobra.Command, name string, short string, value bool, help ...string) *bool {
	helpText := ""
	if len(help) > 0 {
		helpText = help[0]
	}
	return cmd.Flags().BoolP(name, short, value, helpText)
}

// Helper function to attach a string flag
func attachStringFlag(cmd *cobra.Command, name string, short string, value string, help ...string) *string {
	helpText := ""
	if len(help) > 0 {
		helpText = help[0]
	}
	return cmd.Flags().StringP(name, short, value, helpText)
}

// Helper function to attach a string array flag
func attachStringArrayFlag(cmd *cobra.Command, name string, short string, value []string, help ...string) *[]string {
	helpText := ""
	if len(help) > 0 {
		helpText = help[0]
	}
	return cmd.Flags().StringArrayP(name, short, value, helpText)
}
