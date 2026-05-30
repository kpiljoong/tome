package core

import (
	"encoding/json"
	"errors"
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
	uploads           []recordedUpload
	dirs              []recordedUpload
	entries           []*model.JournalEntry
	blobs             map[string][]byte
	blobErrs          map[string]error
	getBlobCalls      []string
	namespaces        []string
	uploadErrs        map[string]error
	listJournalErrs   map[string]error
	listNamespacesErr error
}

func (r *recordingRemote) UploadFile(localPath, remotePath string) error {
	r.uploads = append(r.uploads, recordedUpload{localPath: localPath, remotePath: remotePath})
	if err := r.uploadErrs[remotePath]; err != nil {
		return err
	}
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
	if err := r.listJournalErrs[namespace]; err != nil {
		return nil, err
	}
	return r.entries, nil
}

func (r *recordingRemote) GetBlobByHash(hash string) ([]byte, error) {
	r.getBlobCalls = append(r.getBlobCalls, hash)
	if err := r.blobErrs[hash]; err != nil {
		return nil, err
	}
	return r.blobs[hash], nil
}

func (r *recordingRemote) ListNamespaces() ([]string, error) {
	if r.listNamespacesErr != nil {
		return nil, r.listNamespacesErr
	}
	return r.namespaces, nil
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

func TestSyncNamespaceReturnsErrorWhenBlobUploadFails(t *testing.T) {
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

	remote := &recordingRemote{
		uploadErrs: map[string]error{
			paths.RemoteBlobPath(blobHash): errors.New("remote blob denied"),
		},
	}

	err = SyncNamespace(namespace, remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "namespace sync workbooks completed with 1 blob failure")
	assert.ErrorContains(t, err, "upload blob sha256:abc123")

	require.Len(t, remote.uploads, 1)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
	assert.Empty(t, remote.dirs)
}

func TestSyncNamespaceReturnsErrorWhenJournalCannotBeParsed(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	require.NoError(t, os.WriteFile(paths.JournalPath(namespace, "entry-1"), []byte("{"), 0o644))

	remote := &recordingRemote{}

	err := SyncNamespace(namespace, remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to get referenced blobs")
	assert.ErrorContains(t, err, "parse journal")
	assert.Empty(t, remote.uploads)
	assert.Empty(t, remote.dirs)
}

func TestPullNamespaceFetchesBlobWhenJournalEntryAlreadyExists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:abc123"

	entry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  blobHash,
	}

	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	entryData, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), entryData, 0o644))
	require.NoFileExists(t, paths.BlobPath(blobHash))

	remote := &recordingRemote{
		entries: []*model.JournalEntry{entry},
		blobs: map[string][]byte{
			blobHash: []byte("remote blob"),
		},
	}
	require.NoError(t, PullNamespace(namespace, remote))

	assert.Equal(t, []string{blobHash}, remote.getBlobCalls)
	blobData, err := os.ReadFile(paths.BlobPath(blobHash))
	require.NoError(t, err)
	assert.Equal(t, []byte("remote blob"), blobData)

	storedEntryData, err := os.ReadFile(paths.JournalPath(namespace, entry.ID))
	require.NoError(t, err)
	assert.JSONEq(t, string(entryData), string(storedEntryData))
}

func TestPullNamespaceReturnsErrorWhenBlobFetchFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:missing"

	entry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  blobHash,
	}

	remote := &recordingRemote{
		entries: []*model.JournalEntry{entry},
		blobErrs: map[string]error{
			blobHash: errors.New("remote blob unavailable"),
		},
	}

	err := PullNamespace(namespace, remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "pull namespace workbooks completed with 1 failure")
	assert.ErrorContains(t, err, "fetch blob sha256:missing")
	assert.Equal(t, []string{blobHash}, remote.getBlobCalls)
	require.NoFileExists(t, paths.BlobPath(blobHash))
	require.NoFileExists(t, paths.JournalPath(namespace, entry.ID))
}

func TestSyncBidirectionalReturnsErrorWhenRemoteBlobFetchFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:missing"

	entry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  blobHash,
	}

	remote := &recordingRemote{
		namespaces: []string{namespace},
		entries:    []*model.JournalEntry{entry},
		blobErrs: map[string]error{
			blobHash: errors.New("remote blob unavailable"),
		},
	}

	err := SyncBidirectional(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "bidirectional sync completed with 1 failure")
	assert.ErrorContains(t, err, "fetch blob sha256:missing")
	assert.Equal(t, []string{blobHash}, remote.getBlobCalls)
	require.NoFileExists(t, paths.BlobPath(blobHash))
	require.NoFileExists(t, paths.JournalPath(namespace, entry.ID))
}

func TestSyncBidirectionalReturnsErrorWhenUploadFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:abc123"
	const entryID = "entry-1"

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	require.NoError(t, os.WriteFile(paths.BlobPath(blobHash), []byte("local blob"), 0o644))

	entry := &model.JournalEntry{
		ID:        entryID,
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: namespace,
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  blobHash,
	}
	entryData, err := json.Marshal(entry)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), entryData, 0o644))

	remoteJournalPath := paths.RemoteJournalPath(namespace, entryID)
	remote := &recordingRemote{
		uploadErrs: map[string]error{
			remoteJournalPath: errors.New("remote journal denied"),
		},
	}

	err = SyncBidirectional(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "bidirectional sync completed with 1 failure")
	assert.ErrorContains(t, err, "upload journal")
	assert.ErrorContains(t, err, "remote journal denied")

	require.Len(t, remote.uploads, 2)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
	assert.Equal(t, remoteJournalPath, remote.uploads[1].remotePath)
}

func TestSyncBidirectionalSkipsJournalUploadWhenBlobUploadFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	const blobHash = "sha256:abc123"

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	require.NoError(t, os.WriteFile(paths.BlobPath(blobHash), []byte("local blob"), 0o644))

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

	remote := &recordingRemote{
		uploadErrs: map[string]error{
			paths.RemoteBlobPath(blobHash): errors.New("remote blob denied"),
		},
	}

	err = SyncBidirectional(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "bidirectional sync completed with 1 failure")
	assert.ErrorContains(t, err, "upload blob")
	assert.ErrorContains(t, err, "remote blob denied")

	require.Len(t, remote.uploads, 1)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
}
