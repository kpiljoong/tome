package core

import (
	"fmt"

	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
)

func validateSyncEntry(namespace string, entry *model.JournalEntry) error {
	if entry == nil {
		return fmt.Errorf("namespace=%s sync_phase=validate_entry: nil journal entry", namespace)
	}
	if err := paths.ValidateNamespace(namespace); err != nil {
		return fmt.Errorf("namespace=%s entry_id=%s blob_hash=%s sync_phase=validate_entry: %w", namespace, entry.ID, entry.BlobHash, err)
	}
	if entry.Namespace != "" && entry.Namespace != namespace {
		return fmt.Errorf("namespace=%s entry_id=%s blob_hash=%s sync_phase=validate_entry: entry namespace %q does not match", namespace, entry.ID, entry.BlobHash, entry.Namespace)
	}
	if err := paths.ValidateEntryID(entry.ID); err != nil {
		return fmt.Errorf("namespace=%s entry_id=%s blob_hash=%s sync_phase=validate_entry: %w", namespace, entry.ID, entry.BlobHash, err)
	}
	if err := paths.ValidateBlobHash(entry.BlobHash); err != nil {
		return fmt.Errorf("namespace=%s entry_id=%s blob_hash=%s sync_phase=validate_entry: %w", namespace, entry.ID, entry.BlobHash, err)
	}
	return nil
}
