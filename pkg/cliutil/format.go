package cliutil

import (
	"fmt"

	"github.com/kpiljoong/tome/pkg/model"
)

func FormatEntry(e *model.JournalEntry) string {
	t := e.Timestamp.Format("2006-01-02 15:04")
	return fmt.Sprintf("🧾  %-20s  [%-16s]  ID: %-26s  Blob: %.16s...", e.Filename, t, e.ID, e.BlobHash)
}
