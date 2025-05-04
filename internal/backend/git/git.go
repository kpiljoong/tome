package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
	"github.com/kpiljoong/tome/pkg/paths"
	"github.com/kpiljoong/tome/pkg/util"
)

type GitRepoBackend struct {
	RemoteURL string
	LocalPath string
}

func NewGitRepoBackend(remoteURL string) (*GitRepoBackend, error) {
	cacheDir := filepath.Join(os.TempDir(), "tome-git", util.Slugify(remoteURL))
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); os.IsNotExist(err) {
		logx.Info("📥 Cloning repo: %s → %s", remoteURL, cacheDir)
		cmd := exec.Command("git", "clone", remoteURL, cacheDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git clone failed: %w\n%s", err, string(output))
		}
	} else {
		logx.Info("🔄 Pulling latest: %s", remoteURL)
		cmd := exec.Command("git", "-C", cacheDir, "pull")
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git pull failed: %w\n%s", err, string(output))
		}
	}

	return &GitRepoBackend{
		RemoteURL: remoteURL,
		LocalPath: cacheDir,
	}, nil
}

func (g *GitRepoBackend) UploadDir(localRoot, remoteSubpath string) error {
	dest := filepath.Join(g.LocalPath, remoteSubpath)
	logx.Info("📁 Copying: %s → %s", localRoot, dest)

	if err := util.CopyDir(localRoot, dest); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// Stage all changes
	cmd := exec.Command("git", "-C", g.LocalPath, "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Use `git diff --cached --quiet` to check if anything changed
	check := exec.Command("git", "-C", g.LocalPath, "diff", "--cached", "--quiet")
	if err := check.Run(); err == nil {
		logx.Info("✅ Nothing to sync for: %s", remoteSubpath)
		return nil
	}

	// Something has changed - commit and push
	cmd = exec.Command("git", "-C", g.LocalPath, "commit", "-m", "tome: sync update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	cmd = exec.Command("git", "-C", g.LocalPath, "push")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

func (g *GitRepoBackend) Exists(remotePath string) (bool, error) {
	full := filepath.Join(g.LocalPath, remotePath)
	_, err := os.Stat(full)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (g *GitRepoBackend) ListJournal(namespace, query string) ([]*model.JournalEntry, error) {
	journalDir := filepath.Join(g.LocalPath, "journals", namespace)
	files, err := os.ReadDir(journalDir)
	if err != nil {
		return nil, err
	}

	var results []*model.JournalEntry
	query = strings.ToLower(query)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(journalDir, file.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if query == "" ||
			strings.Contains(strings.ToLower(entry.Filename), query) ||
			strings.Contains(strings.ToLower(entry.FullPath), query) {
			results = append(results, &entry)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

func (g *GitRepoBackend) GetBlobByHash(hash string) ([]byte, error) {
	safeHash := paths.SanitizeHash(hash)
	blobPath := filepath.Join(g.LocalPath, "blobs", safeHash)

	data, err := os.ReadFile(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}
	return data, nil
}

func (g *GitRepoBackend) ListNamespaces() ([]string, error) {
	journalRoot := filepath.Join(g.LocalPath, "journals")  // 👈 not ~/.tome

	logx.Info("Looking for namespaces in: %s", journalRoot)

	entries, err := os.ReadDir(journalRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read journal root: %w", err)
	}

	var namespaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			logx.Info("  found entry: %s (dir=%v)", entry.Name(), entry.IsDir())
			namespaces = append(namespaces, entry.Name())
		}
	}
	return namespaces, nil
}

func (g *GitRepoBackend) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
	// Only support raw GitHub URLs for now
	if !strings.Contains(g.RemoteURL, "github.com") {
		return "", fmt.Errorf("presigned URL generation is only supported for GitHub remotes")
	}

	baseURL := strings.Replace(g.RemoteURL, "github.com", "raw.githubusercontent.com", 1)
	baseURL = strings.TrimSuffix(baseURL, ".git")

	fmt.Printf("=== key === %s\n", key)
	safeKey := paths.SanitizeHash(key)

	url := fmt.Sprintf("%s/refs/heads/main/%s", baseURL, filepath.ToSlash(safeKey))
	return url, nil
}

func (g *GitRepoBackend) BlobKey(hash string) string {
	return filepath.ToSlash(filepath.Join("blobs", paths.SanitizeHash(hash)))
}

func (g *GitRepoBackend) Describe() string {
	return fmt.Sprintf("Git Repo: %s", g.RemoteURL)
}

func (g *GitRepoBackend) UploadFile(localPath, remotePath string) error {
	dest := filepath.Join(g.LocalPath, remotePath)
	logx.Info("📁 Copying: %s → %s", localPath, dest)

	if err := util.CopyFile(localPath, dest); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	cmd := exec.Command("git", "-C", g.LocalPath, "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, string(output))
	}

	cmd = exec.Command("git", "-C", g.LocalPath, "commit", "-m", "tome: sync update")
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=tome", "GIT_COMMIT_NAME=tome")

	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "nothing to commit") {
			// if !contains(string(output), "nothing to commit") {
			return fmt.Errorf("git commit failed: %w\n%s", err, string(output))
		}
		logx.Info("✅ No changes to commit.")
	}

	cmd = exec.Command("git", "-C", g.LocalPath, "push")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, string(output))
	}

	logx.Success("🚀 Synced to Git: %s", g.RemoteURL)
	return nil
}

// func contains(s, substr string) bool {
// 	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && contains(s[1:], substr)))
// }
