package paths

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
