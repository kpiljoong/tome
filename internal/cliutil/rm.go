package cliutil

import "os"

func SafeDelete(path string) error {
	return os.Remove(path)
}
