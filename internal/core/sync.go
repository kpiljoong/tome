package core

import (
	"fmt"
	"os"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/paths"
)

func Sync(localPath string, remote backend.RemoteBackend) error {
	logx.Section("🔄 Syncing blobs...")
	if err := remote.UploadDir(paths.BlobsDir(), paths.RemoteBlobsPrefix); err != nil {
		return fmt.Errorf("blobs sync failed: %w", err)
	}

	namespaces, err := os.ReadDir(paths.JournalsDir())
	if err != nil {
		return fmt.Errorf("failed to read journal directory: %w", err)
	}

	logx.Section("📒 Syncing journal entries...")
	for _, ns := range namespaces {
		if !ns.IsDir() {
			continue
		}
		nsName := ns.Name()
		logx.Info("📂 Namespace: %s", nsName)

		nsPath := paths.NamespaceDir(nsName)
		remotePrefix := paths.RemoteNamespacePrefix(nsName)

		if err := remote.UploadDir(nsPath, remotePrefix); err != nil {
			return fmt.Errorf("sync failed for namespace %s: %w", ns.Name(), err)
		}
	}
	return nil
}
