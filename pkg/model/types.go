package model

import "time"

type JournalEntry struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Namespace string            `json:"namespace"`
	Filename  string            `json:"filename"`
	FullPath  string            `json:"full_path"`
	BlobHash  string            `json:"blob_hash"`
	Meta      map[string]string `json:"meta"`
}
