package core

import (
	"encoding/json"
	"errors"
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
	if err := paths.ValidateNamespace(namespace); err != nil {
		return fmt.Errorf("invalid namespace for sync: %w", err)
	}
	logx.Section("🔄 Syncing namespace: %s", namespace)

	blobRefs, err := getReferencedBlobs(namespace)
	if err != nil {
		return fmt.Errorf("failed to get referenced blobs: %w", err)
	}

	failures := uploadReferencedBlobs(namespace, blobRefs, remote)
	if len(failures) > 0 {
		return fmt.Errorf("namespace sync %s completed with %d blob failure(s): %w", namespace, len(failures), errors.Join(failures...))
	}

	journalPath := paths.NamespaceDir(namespace)
	remotePrefix := paths.RemoteNamespacePrefix(namespace)

	if err := remote.UploadDir(journalPath, remotePrefix); err != nil {
		return fmt.Errorf("failed to upload journal for namespace: %s: %w", namespace, err)
	}

	return nil
}

type referencedBlob struct {
	hash     string
	entryIDs []string
}

func backendBlobOpError(namespace, entryID, blobHash, operation string, err error) error {
	return fmt.Errorf("namespace=%s entry_id=%s blob_hash=%s backend_operation=%s: %w", namespace, entryID, blobHash, operation, err)
}

func backendOperationError(operation string, err error) error {
	return fmt.Errorf("backend_operation=%s: %w", operation, err)
}

func uploadReferencedBlobs(namespace string, blobRefs []referencedBlob, remote backend.RemoteBackend) []error {
	var failures []error
	for _, blobRef := range blobRefs {
		validEntries := true
		for _, entryID := range blobRef.entryIDs {
			if err := paths.ValidateEntryID(entryID); err != nil {
				failures = append(failures, backendBlobOpError(namespace, entryID, blobRef.hash, "ValidateEntryID", err))
				validEntries = false
			}
		}
		if !validEntries {
			continue
		}
		if err := paths.ValidateBlobHash(blobRef.hash); err != nil {
			failures = append(failures, backendBlobOpError(namespace, strings.Join(blobRef.entryIDs, ","), blobRef.hash, "ValidateBlobHash", err))
			continue
		}
		blobPath, _ := paths.SafeBlobPath(blobRef.hash)
		remotePath := paths.RemoteBlobPath(blobRef.hash)

		if err := remote.UploadFile(blobPath, remotePath); err != nil {
			logx.Warn("Failed to upload blob %s: %v", blobRef.hash, err)
			failures = append(failures, backendBlobOpError(namespace, strings.Join(blobRef.entryIDs, ","), blobRef.hash, "UploadFile", err))
		}
	}
	return failures
}

func getReferencedBlobs(namespace string) ([]referencedBlob, error) {
	journalDir := paths.NamespaceDir(namespace)
	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]int)
	var blobRefs []referencedBlob
	var failures []error

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(journalDir, f.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("read journal %s: %w", fullPath, err))
			continue
		}
		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			failures = append(failures, fmt.Errorf("parse journal %s: %w", fullPath, err))
			continue
		}

		hash := entry.BlobHash
		if hash == "" {
			continue
		}
		entryID := entry.ID
		if entryID == "" {
			entryID = strings.TrimSuffix(f.Name(), ".json")
		}
		if idx, ok := seen[hash]; ok {
			blobRefs[idx].entryIDs = append(blobRefs[idx].entryIDs, entryID)
			continue
		}
		seen[hash] = len(blobRefs)
		blobRefs = append(blobRefs, referencedBlob{
			hash:     hash,
			entryIDs: []string{entryID},
		})
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("collect referenced blobs for namespace %s completed with %d failure(s): %w", namespace, len(failures), errors.Join(failures...))
	}
	return blobRefs, nil
}

func Sync(localPath string, remote backend.RemoteBackend) error {
	logx.Section("🔄 Syncing blobs...")
	if err := remote.UploadDir(paths.BlobsDir(), paths.RemoteBlobsPrefix); err != nil {
		return fmt.Errorf("blobs sync failed: %w", backendOperationError("UploadDir", err))
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
			return fmt.Errorf("sync failed for namespace %s: %w", nsName, err)
		}
	}
	return nil
}
