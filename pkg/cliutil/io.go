package cliutil

import (
	"fmt"
	"os"
)

// WriteOutput writes data to a file or stdout.
// If quite is true, it suppresses the success message.
func WriteOutput(path string, data []byte, quiet bool) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	if !quiet {
		fmt.Printf("Output written to: %s\n", path)
	}

	return nil
}
