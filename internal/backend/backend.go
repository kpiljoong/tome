package backend

import "github.com/kpiljoong/tome/pkg/model"

type RemoteBackend interface {
	UploadFile(localPath, remotePath string) error
	Exists(remotePath string) (bool, error)
	UploadDir(localRoot, remotePrefix string) error

	ListJournal(namespace, query string) ([]*model.JournalEntry, error)
	GetBlobByHash(hash string) ([]byte, error)
	ListNamespaces() ([]string, error)

	Describe() string
}
