package cliutil_test

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/kpiljoong/tome/internal/cliutil"
)

func TestFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	cliutil.AttachOutputFlag(rootCmd, "defaultPath")
	cliutil.AttachQuietFlag(rootCmd)
	cliutil.AttachJSONFlag(rootCmd)
	cliutil.AttachLimitFlag(rootCmd)
	cliutil.AttachRemoteFlag(rootCmd, "remote")
	cliutil.AttachModeFlag(rootCmd)
	cliutil.AttachInteractiveFlag(rootCmd)
	cliutil.AttachSmartFlag(rootCmd)
	cliutil.AttachShortenFlag(rootCmd)
	cliutil.AttachExcludeFlag(rootCmd)
	cliutil.AttachAllFlag(rootCmd)

	flags := []string{
		cliutil.FlagOutput,
		cliutil.FlagJSON,
		cliutil.FlagQuiet,
		cliutil.FlagLimit,
		cliutil.FlagMode,
		cliutil.FlagInteractive,
		cliutil.FlagSmart,
		cliutil.FlagShorten,
		cliutil.FlagExclude,
		cliutil.FlagAll,
	}
	for _, f := range flags {
		if rootCmd.Flags().Lookup(f) == nil {
			t.Errorf("Expected flag %s to be registered", f)
		}
	}
}
