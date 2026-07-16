package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
	entriesByNS       map[string][]*model.JournalEntry
	blobs             map[string][]byte
	blobErrs          map[string]error
	getBlobCalls      []string
	namespaces        []string
	uploadErrs        map[string]error
	dirErrs           map[string]error
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
	if err := r.dirErrs[remotePrefix]; err != nil {
		return err
	}
	return nil
}

func (r *recordingRemote) ListJournal(namespace, query string) ([]*model.JournalEntry, error) {
	if err := r.listJournalErrs[namespace]; err != nil {
		return nil, err
	}
	if entries, ok := r.entriesByNS[namespace]; ok {
		return entries, nil
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
	assert.ErrorContains(t, err, "namespace=workbooks")
	assert.ErrorContains(t, err, "entry_id=entry-1")
	assert.ErrorContains(t, err, "blob_hash=sha256:abc123")
	assert.ErrorContains(t, err, "backend_operation=UploadFile")

	require.Len(t, remote.uploads, 1)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
	assert.Empty(t, remote.dirs)
}

func TestSyncNamespaceAggregatesBlobUploadFailuresWithEntryContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	entries := []*model.JournalEntry{
		{
			ID:        "entry-1",
			Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "plan.json",
			FullPath:  "/tmp/plan.json",
			BlobHash:  "sha256:abc123",
		},
		{
			ID:        "entry-2",
			Timestamp: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "notes.json",
			FullPath:  "/tmp/notes.json",
			BlobHash:  "sha256:def456",
		},
	}

	require.NoError(t, paths.EnsureDirExists(paths.BlobsDir()))
	require.NoError(t, paths.EnsureDirExists(paths.NamespaceDir(namespace)))
	for _, entry := range entries {
		require.NoError(t, os.WriteFile(paths.BlobPath(entry.BlobHash), []byte("blob"), 0o644))
		entryData, err := json.Marshal(entry)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(paths.JournalPath(namespace, entry.ID), entryData, 0o644))
	}

	remote := &recordingRemote{
		uploadErrs: map[string]error{
			paths.RemoteBlobPath("sha256:abc123"): errors.New("first blob denied"),
			paths.RemoteBlobPath("sha256:def456"): errors.New("second blob denied"),
		},
	}

	err := SyncNamespace(namespace, remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "namespace sync workbooks completed with 2 blob failure")
	assert.ErrorContains(t, err, "namespace=workbooks entry_id=entry-1 blob_hash=sha256:abc123 backend_operation=UploadFile")
	assert.ErrorContains(t, err, "namespace=workbooks entry_id=entry-2 blob_hash=sha256:def456 backend_operation=UploadFile")
	assert.ErrorContains(t, err, "first blob denied")
	assert.ErrorContains(t, err, "second blob denied")

	require.Len(t, remote.uploads, 2)
	assert.Empty(t, remote.dirs)
}

func TestSyncReturnsErrorWhenBlobUploadDirFailsWithOperationContext(t *testing.T) {
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
		dirErrs: map[string]error{
			paths.RemoteBlobsPrefix: errors.New("blob dir upload denied"),
		},
	}

	err = Sync(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "blobs sync failed")
	assert.ErrorContains(t, err, "backend_operation=UploadDir")
	assert.ErrorContains(t, err, "blob dir upload denied")

	require.Len(t, remote.dirs, 1)
	assert.Equal(t, paths.BlobsDir(), remote.dirs[0].localPath)
	assert.Equal(t, paths.RemoteBlobsPrefix, remote.dirs[0].remotePath)
	assert.Empty(t, remote.uploads)
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
	assert.ErrorContains(t, err, "namespace=workbooks")
	assert.ErrorContains(t, err, "entry_id=entry-1")
	assert.ErrorContains(t, err, "blob_hash=sha256:missing")
	assert.ErrorContains(t, err, "backend_operation=GetBlobByHash")
	assert.Equal(t, []string{blobHash}, remote.getBlobCalls)
	require.NoFileExists(t, paths.BlobPath(blobHash))
	require.NoFileExists(t, paths.JournalPath(namespace, entry.ID))
}

func TestPullNamespaceAggregatesBlobFetchFailuresWithEntryContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const namespace = "workbooks"
	entries := []*model.JournalEntry{
		{
			ID:        "entry-1",
			Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "plan.json",
			FullPath:  "/tmp/plan.json",
			BlobHash:  "sha256:missing-a",
		},
		{
			ID:        "entry-2",
			Timestamp: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
			Namespace: namespace,
			Filename:  "notes.json",
			FullPath:  "/tmp/notes.json",
			BlobHash:  "sha256:missing-b",
		},
	}

	remote := &recordingRemote{
		entries: entries,
		blobErrs: map[string]error{
			"sha256:missing-a": errors.New("first blob unavailable"),
			"sha256:missing-b": errors.New("second blob unavailable"),
		},
	}

	err := PullNamespace(namespace, remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "pull namespace workbooks completed with 2 failure")
	assert.ErrorContains(t, err, "namespace=workbooks entry_id=entry-1 blob_hash=sha256:missing-a backend_operation=GetBlobByHash")
	assert.ErrorContains(t, err, "namespace=workbooks entry_id=entry-2 blob_hash=sha256:missing-b backend_operation=GetBlobByHash")
	assert.ErrorContains(t, err, "first blob unavailable")
	assert.ErrorContains(t, err, "second blob unavailable")
	assert.Equal(t, []string{"sha256:missing-a", "sha256:missing-b"}, remote.getBlobCalls)
	for _, entry := range entries {
		require.NoFileExists(t, paths.BlobPath(entry.BlobHash))
		require.NoFileExists(t, paths.JournalPath(namespace, entry.ID))
	}
}

func TestPullAggregatesNamespaceBlobFetchFailuresWithEntryContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	workbookEntry := &model.JournalEntry{
		ID:        "entry-1",
		Timestamp: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
		Namespace: "workbooks",
		Filename:  "plan.json",
		FullPath:  "/tmp/plan.json",
		BlobHash:  "sha256:missing-a",
	}
	noteEntry := &model.JournalEntry{
		ID:        "entry-2",
		Timestamp: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		Namespace: "notes",
		Filename:  "today.json",
		FullPath:  "/tmp/today.json",
		BlobHash:  "sha256:missing-b",
	}

	remote := &recordingRemote{
		namespaces: []string{"workbooks", "notes"},
		entriesByNS: map[string][]*model.JournalEntry{
			"workbooks": {workbookEntry},
			"notes":     {noteEntry},
		},
		blobErrs: map[string]error{
			"sha256:missing-a": errors.New("first pull blob unavailable"),
			"sha256:missing-b": errors.New("second pull blob unavailable"),
		},
	}

	err := Pull(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "pull completed with 2 failure")
	assert.ErrorContains(t, err, "namespace=workbooks entry_id=entry-1 blob_hash=sha256:missing-a backend_operation=GetBlobByHash")
	assert.ErrorContains(t, err, "namespace=notes entry_id=entry-2 blob_hash=sha256:missing-b backend_operation=GetBlobByHash")
	assert.ErrorContains(t, err, "first pull blob unavailable")
	assert.ErrorContains(t, err, "second pull blob unavailable")
	assert.Equal(t, []string{"sha256:missing-a", "sha256:missing-b"}, remote.getBlobCalls)
	require.NoFileExists(t, paths.BlobPath(workbookEntry.BlobHash))
	require.NoFileExists(t, paths.JournalPath(workbookEntry.Namespace, workbookEntry.ID))
	require.NoFileExists(t, paths.BlobPath(noteEntry.BlobHash))
	require.NoFileExists(t, paths.JournalPath(noteEntry.Namespace, noteEntry.ID))
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
	assert.ErrorContains(t, err, "namespace=workbooks")
	assert.ErrorContains(t, err, "entry_id=entry-1")
	assert.ErrorContains(t, err, "blob_hash=sha256:missing")
	assert.ErrorContains(t, err, "backend_operation=GetBlobByHash")
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
	assert.ErrorContains(t, err, "namespace=workbooks")
	assert.ErrorContains(t, err, "entry_id=entry-1")
	assert.ErrorContains(t, err, "blob_hash=sha256:abc123")
	assert.ErrorContains(t, err, "backend_operation=UploadFile")
	assert.ErrorContains(t, err, "remote blob denied")

	require.Len(t, remote.uploads, 1)
	assert.Equal(t, paths.RemoteBlobPath(blobHash), remote.uploads[0].remotePath)
}

func TestPullRejectsTraversalNamespaceBeforeMaterializingRemoteData(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	remote := &recordingRemote{namespaces: []string{"../outside"}}
	err := Pull(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid namespace path segment")
	require.NoDirExists(t, filepath.Join(home, "outside"))
}

func TestPullNamespaceRejectsTraversalEntryIDBeforeWritingJournal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	entry := &model.JournalEntry{
		ID:        "../outside",
		Namespace: "workbooks",
		Filename:  "plan.json",
		BlobHash:  "sha256:abc123",
	}
	remote := &recordingRemote{entries: []*model.JournalEntry{entry}}

	err := PullNamespace("workbooks", remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "entry_id=../outside")
	assert.ErrorContains(t, err, "invalid entry ID path segment")
	assert.Empty(t, remote.getBlobCalls)
	require.NoFileExists(t, filepath.Join(paths.JournalsDir(), "outside.json"))
}

func TestPullNamespaceRejectsTraversalBlobHashBeforeWritingBlob(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	entry := &model.JournalEntry{
		ID:        "entry-1",
		Namespace: "workbooks",
		Filename:  "plan.json",
		BlobHash:  "../outside",
	}
	remote := &recordingRemote{entries: []*model.JournalEntry{entry}}

	err := PullNamespace("workbooks", remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "blob_hash=../outside")
	assert.ErrorContains(t, err, "invalid blob hash path segment")
	assert.Empty(t, remote.getBlobCalls)
	require.NoFileExists(t, filepath.Join(paths.TomeRoot(), "outside"))
}

func TestSyncBidirectionalRejectsMismatchedRemoteEntryNamespace(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	remote := &recordingRemote{
		namespaces: []string{"workbooks"},
		entries: []*model.JournalEntry{{
			ID:        "entry-1",
			Namespace: "notes",
			Filename:  "plan.json",
			BlobHash:  "sha256:abc123",
		}},
	}

	err := SyncBidirectional(paths.TomeRoot(), remote)
	require.Error(t, err)
	assert.ErrorContains(t, err, "entry namespace \"notes\" does not match")
	assert.Empty(t, remote.getBlobCalls)
	require.NoFileExists(t, paths.JournalPath("workbooks", "entry-1"))
}
