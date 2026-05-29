package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

func SyncNamespace(namespace string, remote backend.RemoteBackend) error {
	logx.Section("🔄 Syncing namespace: %s", namespace)

	blobIDs, err := getReferencedBlobs(namespace)
	if err != nil {
		return fmt.Errorf("failed to get referenced blobs: %w", err)
	}

	for _, blobID := range blobIDs {
		blobPath := paths.BlobPath(blobID)
		remotePath := paths.RemoteBlobPath(blobID)

		if err := remote.UploadFile(blobPath, remotePath); err != nil {
			logx.Warn("Failed to upload blob %s: %v", blobID, err)
		}
	}

	journalPath := paths.NamespaceDir(namespace)
	remotePrefix := paths.RemoteNamespacePrefix(namespace)

	if err := remote.UploadDir(journalPath, remotePrefix); err != nil {
		return fmt.Errorf("failed to upload journal for namespace: %s: %w", namespace, err)
	}

	return nil
}

func getReferencedBlobs(namespace string) ([]string, error) {
	journalDir := paths.NamespaceDir(namespace)
	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var blobIDs []string

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(journalDir, f.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			logx.Warn("Faeild to read journal: %s", fullPath)
			continue
		}
		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			logx.Warn("Failed to parse journal: %s", fullPath)
			continue
		}

		hash := entry.BlobHash
		if hash != "" && !seen[hash] {
			fmt.Printf("===== Found blob: %s\n", hash)
			seen[hash] = true
			blobIDs = append(blobIDs, hash)
		}
	}
	return blobIDs, nil
}

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
