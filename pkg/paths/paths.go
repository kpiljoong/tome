package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	RemoteJournalsPrefix = "journals"
	RemoteBlobsPrefix    = "blobs"
)

var tomeRoot = filepath.Join(os.Getenv("HOME"), ".tome")

func SetRoot(path string) {
	tomeRoot = path
}

func TomeRoot() string {
	hoem, err := os.UserHomeDir()
	if err != nil {
		panic("could not resolve $HOME")
	}
	return filepath.Join(hoem, ".tome")
}

func BlobsDir() string {
	return filepath.Join(TomeRoot(), RemoteBlobsPrefix)
}

func BlobPath(hash string) string {
	return filepath.Join(BlobsDir(), hash)
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

func EnsureDirExists(path string) error {
	return os.MkdirAll(path, 0o755)
}
