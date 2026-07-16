package git

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
	"github.com/kpiljoong/tome/internal/util"
)

type GitRepoBackend struct {
	RemoteURL string
	LocalPath string
}

func NewGitRepoBackend(remoteURL string) (*GitRepoBackend, error) {
	cacheKey := fmt.Sprintf("%x", sha256.Sum256([]byte(remoteURL)))
	cacheDir := filepath.Join(os.TempDir(), "tome-git", cacheKey)
	if _, err := os.Stat(filepath.Join(cacheDir, ".git")); os.IsNotExist(err) {
		logx.Info("📥 Cloning repo: %s → %s", remoteURL, cacheDir)
		cmd := exec.Command("git", "clone", remoteURL, cacheDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git clone failed: %w\n%s", err, string(output))
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to inspect git cache: %w", err)
	} else {
		origin, err := exec.Command("git", "-C", cacheDir, "remote", "get-url", "origin").CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to read cached remote origin: %w\n%s", err, string(origin))
		}
		if strings.TrimSpace(string(origin)) != remoteURL {
			return nil, fmt.Errorf("cached remote origin does not match requested remote")
		}
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
	cmd = gitCommitCommand(g.LocalPath)
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
	if err := paths.ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	journalDir := filepath.Join(g.LocalPath, "journals", namespace)
	files, err := os.ReadDir(journalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var results []*model.JournalEntry
	var failures []error
	query = strings.ToLower(query)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(journalDir, file.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			failures = append(failures, fmt.Errorf("read journal %s: %w", fullPath, err))
			continue
		}

		var entry model.JournalEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			failures = append(failures, fmt.Errorf("decode journal %s: %w", fullPath, err))
			continue
		}

		if query == "" ||
			strings.Contains(strings.ToLower(entry.Filename), query) ||
			strings.Contains(strings.ToLower(entry.FullPath), query) {
			results = append(results, &entry)
		}
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("list git journal %s completed with %d failure(s): %w", namespace, len(failures), errors.Join(failures...))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	return results, nil
}

func (g *GitRepoBackend) GetBlobByHash(hash string) ([]byte, error) {
	if err := paths.ValidateBlobHash(hash); err != nil {
		return nil, fmt.Errorf("invalid blob hash: %w", err)
	}
	safeHash := paths.SanitizeHash(hash)
	blobPath := filepath.Join(g.LocalPath, "blobs", safeHash)

	data, err := os.ReadFile(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}
	return data, nil
}

func (g *GitRepoBackend) ListNamespaces() ([]string, error) {
	journalRoot := filepath.Join(g.LocalPath, "journals")

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

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}
	if err := util.CopyFile(localPath, dest); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	cmd := exec.Command("git", "-C", g.LocalPath, "add", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, string(output))
	}

	cmd = gitCommitCommand(g.LocalPath)

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

func gitCommitCommand(repoPath string) *exec.Cmd {
	cmd := exec.Command("git", "-C", repoPath, "commit", "-m", "tome: sync update")
	cmd.Env = gitIdentityEnv()
	return cmd
}

func gitIdentityEnv() []string {
	identityKeys := []string{
		"GIT_AUTHOR_NAME=",
		"GIT_AUTHOR_EMAIL=",
		"GIT_COMMITTER_NAME=",
		"GIT_COMMITTER_EMAIL=",
	}

	var env []string
	for _, value := range os.Environ() {
		if !hasAnyPrefix(value, identityKeys) {
			env = append(env, value)
		}
	}
	return append(env,
		"GIT_AUTHOR_NAME=tome",
		"GIT_AUTHOR_EMAIL=tome@localhost",
		"GIT_COMMITTER_NAME=tome",
		"GIT_COMMITTER_EMAIL=tome@localhost",
	)
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

// func contains(s, substr string) bool {
// 	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && contains(s[1:], substr)))
// }
