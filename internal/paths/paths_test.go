package paths

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafePathHelpersRejectTraversal(t *testing.T) {
	for _, value := range []string{"", ".", "..", "../outside", "nested/name", `nested\\name`, "/tmp/outside"} {
		t.Run(value, func(t *testing.T) {
			assert.Error(t, ValidateNamespace(value))
			assert.Error(t, ValidateEntryID(value))
			assert.Error(t, ValidateBlobHash(value))
		})
	}

	assert.Error(t, ValidateEntryID("entry.json"))
}

func TestSafePathHelpersPreserveValidPaths(t *testing.T) {
	root := t.TempDir()
	SetRoot(root)
	t.Cleanup(func() { SetRoot("") })

	journalPath, err := SafeJournalPath("workbooks", "entry-1")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "journals", "workbooks", "entry-1.json"), journalPath)

	blobPath, err := SafeBlobPath("sha256:abc123")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "blobs", "sha256_abc123"), blobPath)
}

func TestTomeRootUsesHomeWhenNoOverride(t *testing.T) {
	SetRoot("")
	home := t.TempDir()
	t.Setenv("HOME", home)

	assert.Equal(t, filepath.Join(home, ".tome"), TomeRoot())
}

func TestSetRootOverridesTomeRoot(t *testing.T) {
	root := t.TempDir()
	SetRoot(root)
	t.Cleanup(func() {
		SetRoot("")
	})

	assert.Equal(t, root, TomeRoot())
}
