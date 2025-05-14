package cliutil_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kpiljoong/tome/internal/cliutil"
	"github.com/kpiljoong/tome/internal/model"
)

func TestFormatEntry(t *testing.T) {
	timestamp, _ := time.Parse("2006-01-02 15:04", "2023-10-01 12:00")
	entry := &model.JournalEntry{
		Filename:  "test.txt",
		Timestamp: timestamp,
		ID:        "1234567890abcdef1234567890abcdef",
		BlobHash:  "abcdef1234567890abcdef1234567890",
	}

	expectedValues := []string{
		entry.Filename,
		entry.Timestamp.Format("2006-01-02 15:04"),
		entry.ID,
		strings.SplitN(entry.ID, " ", 10)[0],
	}
	result := cliutil.FormatEntry(entry)
	for _, expected := range expectedValues {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected '%s' to contain '%s'", result, expected)
		}
	}
}
