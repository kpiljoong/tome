package core

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedUpload struct {
	localPath  string
	remotePath string
}

type recordingRemote struct {
	uploads []recordedUpload
	dirs    []recordedUpload
}

func (r *recordingRemote) UploadFile(localPath, remotePath string) error {
	r.uploads = append(r.uploads, recordedUpload{localPath: localPath, remotePath: remotePath})
	return nil
}

func (r *recordingRemote) Exists(remotePath string) (bool, error) {
	return false, nil
}

func (r *recordingRemote) UploadDir(localRoot, remotePrefix string) error {
	r.dirs = append(r.dirs, recordedUpload{localPath: localRoot, remotePath: remotePrefix})
	return nil
}

func (r *recordingRemote) ListJournal(namespace, query string) ([]*model.JournalEntry, error) {
	return nil, nil
}

func (r *recordingRemote) GetBlobByHash(hash string) ([]byte, error) {
	return nil, nil
}

func (r *recordingRemote) ListNamespaces() ([]string, error) {
	return nil, nil
}

func (r *recordingRemote) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
	return "", nil
}

func (r *recordingRemote) BlobKey(hash string) string {
	return paths.RemoteBlobPath(hash)
}

func (r *recordingRemote) Describe() string {
	return "recording"
}

func TestSyncNamespaceUsesSanitizedRemoteBlobPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:abc123"

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	require.NoError(t, os.WriteFile(paths.BlobPath(blobHash), []byte("blob"), 0o644))

	entry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  blobHash,
	}
	entryData, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), entryData, 0o644))

	remote := &recordingRemote{}
	require.NoError(t, SyncNamespace(namespace, remote))

	require.Len(t, remote.uploads, 1)
	assert.Equal(t, paths.BlobPath(blobHash), remote.uploads[0].localPath)
	assert.Equal(t, "blobs/sha256_abc123", remote.uploads[0].remotePath)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
	assert.NotContains(t, remote.uploads[0].remotePath, ":")

	require.Len(t, remote.dirs, 1)
	assert.Equal(t, paths.NamespaceDir(namespace), remote.dirs[0].localPath)
	assert.Equal(t, paths.RemoteNamespacePrefix(namespace), remote.dirs[0].remotePath)
}
