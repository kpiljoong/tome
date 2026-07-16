package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListJournalReturnsEmptyForMissingNamespace(t *testing.T) {
	backend := &GitRepoBackend{LocalPath: t.TempDir()}

	entries, err := backend.ListJournal("workbooks", "")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestGitBackendRejectsTraversalPaths(t *testing.T) {
	backend := &GitRepoBackend{LocalPath: t.TempDir()}

	_, err := backend.ListJournal("../outside", "")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid namespace")

	_, err = backend.GetBlobByHash("../outside")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid blob hash")
}

func TestListJournalReturnsErrorForDecodeFailure(t *testing.T) {
	root := t.TempDir()
	journalDir := filepath.Join(root, "journals", "workbooks")
	require.NoError(t, os.MkdirAll(journalDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(journalDir, "bad.json"), []byte("{"), 0o644))

	backend := &GitRepoBackend{LocalPath: root}

	entries, err := backend.ListJournal("workbooks", "")
	require.Error(t, err)
	assert.Nil(t, entries)
	assert.ErrorContains(t, err, "list git journal workbooks completed with 1 failure")
	assert.ErrorContains(t, err, "decode journal")
}

func TestListJournalFiltersAndSortsEntries(t *testing.T) {
	root := t.TempDir()
	journalDir := filepath.Join(root, "journals", "workbooks")
	require.NoError(t, os.MkdirAll(journalDir, 0o755))

	older := model.JournalEntry{
		ID:        "older",
		Timestamp: time.Date(2026, 5, 30, 1, 0, 0, 0, time.UTC),
		Namespace: "workbooks",
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  "sha256:old",
	}
	newer := model.JournalEntry{
		ID:        "newer",
		Timestamp: time.Date(2026, 5, 30, 2, 0, 0, 0, time.UTC),
		Namespace: "workbooks",
		Filename:  "plan-final.json",
		FullPath:  "/tmp/plan-final.json",
		BlobHash:  "sha256:new",
	}
	writeJournalEntry(t, filepath.Join(journalDir, "older.json"), older)
	writeJournalEntry(t, filepath.Join(journalDir, "newer.json"), newer)

	backend := &GitRepoBackend{LocalPath: root}

	entries, err := backend.ListJournal("workbooks", "plan")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "newer", entries[0].ID)
	assert.Equal(t, "older", entries[1].ID)
}

func writeJournalEntry(t *testing.T, path string, entry model.JournalEntry) {
	t.Helper()

	data, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o644))
}
