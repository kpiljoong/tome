# Tome

Tome is a zero-runtime, local-first journaling and blob store for developers and systems.
It tracks immutable snapshots of files and lets you:

- Save any file
- Search and retrieve by name
- Sync to or from remote backends like S3 or GitHub
- Serve logs via optional HTTP API
- Compare local vs remote with `status` and `sync`

## Why Tome?

| Feature              | Benefit                                      |
|----------------------|----------------------------------------------|
| Immutable blobs      | Version-safe, content-addressed storage      |
| Journaling           | Metadata: full path, size, mtime, hash       |
| Simple CLI           | `save`, `search`, `get`, `sync`, `status`    |
| Remote support       | Works with S3 and (soon) GitHub              |
| No runtime required  | No daemon, DB, or server needed              |
| Portable and reliable| Good for CI, workflows, audits, archiving    |

## Installation

```bash
go install github.com/kpiljoong/tome@latest
```
