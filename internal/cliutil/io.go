package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kpiljoong/tome/internal/logx"
)

// WriteOutput writes data to a file or stdout.
// If quite is true, it suppresses the success message.
// If asJSON is true, the data is formatted before writing.
func WriteOutput(path string, data []byte, quiet bool, asJSON bool) error {
	if path == "" {
		if asJSON {
			var parsed any
			if err := json.Unmarshal(data, &parsed); err != nil {
				return fmt.Errorf("failed to parse JSON data: %w", err)
			}
			if !quiet {
				logx.Info("📝 Outputting JSON to stdout:")
			}
			return PrintPrettyJSON(parsed)
		}

		if !quiet {
			logx.Info("📝 Outputting raw to stdout:")
		}
		_, err := os.Stdout.Write(data)
		fmt.Println()
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	if !quiet {
		logx.Success("📄 File written to %s", path)
	}

	return nil
}

func PrintPrettyJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
