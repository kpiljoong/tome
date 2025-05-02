package github

import (
	"os"
	"path/filepath"
)

type GitHubBackend struct {
	RepoURL    string
	LocalClone string
}

func NewGitHubBackend(repoURL string) (*GitHubBackend, error) {
	return &GitHubBackend{
		RepoURL:    repoURL,
		LocalClone: filepath.Join(os.TempDir(), "tome-github"),
	}, nil
}

func (g *GitHubBackend) UploadDir(localRoot, remotePrefix string) error {
	return nil
}
