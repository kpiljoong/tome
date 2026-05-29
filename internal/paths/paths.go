package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/internal/model"
)

const (
	RemoteJournalsPrefix = "journals"
	RemoteBlobsPrefix    = "blobs"
)

var tomeRootOverride string

func SetRoot(path string) {
	tomeRootOverride = path
}

func TomeRoot() string {
	if tomeRootOverride != "" {
		return tomeRootOverride
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic("could not resolve $HOME")
	}
	return filepath.Join(home, ".tome")
}

func BlobsDir() string {
	return filepath.Join(TomeRoot(), RemoteBlobsPrefix)
}

func SanitizeHash(hash string) string {
	// Replace ':' with '_' to avoid issues with file paths
	return strings.ReplaceAll(hash, ":", "_")
}

func BlobPath(hash string) string {
	return filepath.Join(BlobsDir(), SanitizeHash(hash))
}

func JournalsDir() string {
	return filepath.Join(TomeRoot(), RemoteJournalsPrefix)
}

func NamespaceDir(ns string) string {
	return filepath.Join(JournalsDir(), ns)
}

func JournalPath(ns, id string) string {
	return filepath.Join(NamespaceDir(ns), fmt.Sprintf("%s.json", id))
}

func JournalEntryPath(entry *model.JournalEntry) string {
	return filepath.Join(JournalsDir(), entry.Namespace, entry.ID+".json")
}

func EnsureDirExists(path string) error {
	return os.MkdirAll(path, 0o755)
}
