package cliutil

import (
	"fmt"
	"strings"

	"github.com/kpiljoong/tome/internal/backend"
	"github.com/kpiljoong/tome/internal/backend/s3"
)

func ResolveRemote(target string, fallback string) (backend.RemoteBackend, error) {
	if target == "" {
		target = fallback
	}
	if target == "" {
		return nil, fmt.Errorf("no remote provided and no fallback")
	}

	switch {
	case strings.HasPrefix(target, "s3://"):
		parts := strings.SplitN(strings.TrimPrefix(target, "s3://"), "/", 2)
		bucket := parts[0]
		prefix := ""
		if len(parts) > 1 {
			prefix = parts[1]
		}
		return s3.NewS3Backend(bucket, prefix)
	default:
		return nil, fmt.Errorf("unsupported backend scheme: %s", target)
	}
}
