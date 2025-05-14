package backend

import (
	"time"

	"github.com/kpiljoong/tome/internal/model"
)

type RemoteBackend interface {
	UploadFile(localPath, remotePath string) error
	Exists(remotePath string) (bool, error)
	UploadDir(localRoot, remotePrefix string) error

	ListJournal(namespace, query string) ([]*model.JournalEntry, error)
	GetBlobByHash(hash string) ([]byte, error)
	ListNamespaces() ([]string, error)

	GeneratePresignedURL(key string, expiry time.Duration) (string, error)

	BlobKey(hash string) string
	Describe() string
}
