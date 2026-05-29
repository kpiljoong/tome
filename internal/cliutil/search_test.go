package cliutil_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalSearchUsesExactFilenameOrFullPathMatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	entries := []*model.JournalEntry{
		{
			ID:        "entry-1",
			Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "plan.json",
			FullPath:  "/tmp/plan.json",
			BlobHash:  "sha256:one",
		},
		{
			ID:        "entry-2",
			Timestamp: time.Date(2026, 5, 29, 0, 1, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "plan-final.json",
			FullPath:  "/tmp/plan-final.json",
			BlobHash:  "sha256:two",
		},
	}

	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), data, 0o644))
	}

	matches, err := cliutil.LocalSearch(namespace, "plan")
	require.NoError(t, err)
	assert.Empty(t, matches)

	matches, err = cliutil.LocalSearch(namespace, "PLAN.JSON")
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "entry-1", matches[0].ID)

	matches, err = cliutil.LocalSearch(namespace, "/tmp/plan-final.json")
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "entry-2", matches[0].ID)

	matches, err = cliutil.LocalSearch(namespace, "")
	require.NoError(t, err)
	assert.Len(t, matches, 2)
}
