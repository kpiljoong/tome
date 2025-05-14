package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/stretchr/testify/assert"
)

func TestSave_WritesBlobAndJournal(t *testing.T) {
	tmpDir := t.TempDir()
	paths.SetRoot(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("this is a test")
	err := os.WriteFile(testFile, content, 0o644)
	assert.NoError(t, err)

	entry, err := Save("testspace", testFile, false)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "test.txt", entry.Filename)
	assert.Contains(t, entry.BlobHash, "sha256:")

	blobPath := paths.BlobPath(entry.BlobHash)
	blobData, err := os.ReadFile(blobPath)
	assert.NoError(t, err)
	assert.Equal(t, content, blobData)

	journalPath := paths.JournalPath("testspace", entry.ID)
	journalData, err := os.ReadFile(journalPath)
	assert.NoError(t, err)

	var stored model.JournalEntry
	err = json.Unmarshal(journalData, &stored)
	assert.NoError(t, err)
	assert.Equal(t, entry.ID, stored.ID)
	assert.Equal(t, entry.Filename, stored.Filename)
	assert.Equal(t, entry.BlobHash, stored.BlobHash)
}
