package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/pkg/paths"
)

func Sync(localPath string, remote backend.RemoteBackend) error {
	// fmt.Println("Syncing blobs...")
	if err := remote.UploadDir(filepath.Join(localPath, "blobs"), "blobs"); err != nil {
		return fmt.Errorf("blobs sync failed: %w", err)
	}

	// fmt.Println("Syncing journal entries...")
	journalRoot := filepath.Join(localPath, "journals")
	entries, _ := os.ReadDir(journalRoot)

	for _, ns := range entries {
		nsPath := paths.NamespaceDir(ns.Name())
		// nsPath := filepath.Join(journalRoot, ns.Name())
		if err := remote.UploadDir(nsPath, filepath.Join("journals", ns.Name())); err != nil {
			return fmt.Errorf("sync failed for namespace %s: %w", ns.Name(), err)
		}
	}
	return nil
}
