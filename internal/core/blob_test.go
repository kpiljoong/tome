package core

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlobByHashUsesSanitizedBlobPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	hash := "sha256:abc123"
	expected := []byte("stored blob")

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, os.WriteFile(paths.BlobPath(hash), expected, 0o644))

	actual, err := GetBlobByHash(hash)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetUsesSanitizedBlobPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	hash := "sha256:def456"
	expected := []byte("journal blob")
	entry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  hash,
	}

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	require.NoError(t, os.WriteFile(paths.BlobPath(hash), expected, 0o644))

	journalData, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), journalData, 0o644))

	actual, err := Get(namespace, "plan.json")
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
