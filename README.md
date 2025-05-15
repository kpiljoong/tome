# Tome

![Badge](https://hitscounter.dev/api/hit?url=https%3A%2F%2Fgithub.com%2Fkpiljoong%2Ftome&label=Visitor&icon=github&color=%23198754)

**Tome** is a zero-runtime, append-only journal and blob store for files. Designed for developers and systems alike, it lets you save, search, sync, and share file snapshots efficiently — both locally and remotely (e.g., S3, GitHub).

## Features

- **Snapshot** files into an append-only journal
- **Search** by filename or full path
- **Sync** with remote storage (S3, GitHub support)
- **Share** files with temoporary or shortened URLs
- **Organize** with namespaces (like Git branches)
- Minimal dependencies, portable CLI tool

## Installation

```bash
go install github.com/kpiljoong/tome@latest
```

Or download binaries from [Releases](https://github.com/kpiljoong/tome/releases) for Windows, macOS, and Linux.

## Usage

### Save a file

```bash
tome save workbooks plan.json
```

### Search entries

```bash
tome search workbooks plan
```

### List namespace entries

```bash
tome ls workbooks
```

### Get the latest version

```bash
tome latest workbooks plan.json --output ./restored.json
```

### Sync with S3

```bash
tome sync --to s3://your-bucket/prefix
```

### Status check

```bash
tome status --from s3://your-bucket/prefix --json
```

### Share a file

```bash
tome share workbooks plan.json --from s3://your-bucket/prefix --shorten
```

### Terminal UI (TUI)

You can browse saved journal entries in a terminal interface:

```bash
tome tui
```

## Configuration

Create a config file at `~/.tome/config.yaml`:

```yaml
default_remote: s3://your-bucket/prefix
```

## Structure

- `.tome/` - Local store
  - `blobs/` - Content-addressed file blobs
  - `journals/<namespace>/` - Journal entries as JSON
- Remotes (S3/GitHub) mirror the same layout

## Testing

```bash
go test ./...
```

## License

[![MIT](https://img.shields.io/badge/license-MIT-blue)](https://github.com/kpiljoong/tome/blob/master/LICENSE)
