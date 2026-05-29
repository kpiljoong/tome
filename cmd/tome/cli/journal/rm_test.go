package journal

import (
	"testing"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRmArgsAllowsNamespaceOnlyWithAll(t *testing.T) {
	cmd := &cobra.Command{Use: "rm"}
	cliutil.AttachAllFlag(cmd)

	require.NoError(t, cmd.Flags().Set(cliutil.FlagAll, "true"))
	assert.NoError(t, validateRmArgs(cmd, []string{"workbooks"}))
	assert.NoError(t, validateRmArgs(cmd, []string{"workbooks", "plan.json"}))
	assert.Error(t, validateRmArgs(cmd, nil))
	assert.Error(t, validateRmArgs(cmd, []string{"workbooks", "plan.json", "extra"}))

	require.NoError(t, cmd.Flags().Set(cliutil.FlagAll, "false"))
	assert.NoError(t, validateRmArgs(cmd, []string{"workbooks", "plan.json"}))
	assert.Error(t, validateRmArgs(cmd, []string{"workbooks"}))
}

func TestResolveRmTargetsConfirmsNamespaceWideAllBeforeSingleFastPath(t *testing.T) {
	entries := []*model.JournalEntry{{ID: "entry-1", Filename: "plan.json"}}
	confirmed := false

	targets, aborted, err := resolveRmTargets(
		"workbooks",
		"",
		entries,
		true,
		false,
		func() bool {
			confirmed = true
			return false
		},
		nil,
	)

	require.NoError(t, err)
	assert.True(t, confirmed)
	assert.True(t, aborted)
	assert.Nil(t, targets)
}

func TestResolveRmTargetsQueryAllDoesNotRequireNamespaceWideConfirmation(t *testing.T) {
	entries := []*model.JournalEntry{
		{ID: "entry-1", Filename: "plan.json"},
		{ID: "entry-2", Filename: "plan.json"},
	}
	confirmed := false

	targets, aborted, err := resolveRmTargets(
		"workbooks",
		"plan.json",
		entries,
		true,
		false,
		func() bool {
			confirmed = true
			return false
		},
		nil,
	)

	require.NoError(t, err)
	assert.False(t, confirmed)
	assert.False(t, aborted)
	assert.Equal(t, entries, targets)
}
