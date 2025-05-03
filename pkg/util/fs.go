package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		return CopyFile(path, destPath)
	})
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func ShouldExclude(path string, patterns []string) bool {
	for _, pat := range patterns {
		match, _ := filepath.Match(pat, filepath.Base(path))
		if match || strings.Contains(path, pat) {
			return true
		}
	}
	return false
}
