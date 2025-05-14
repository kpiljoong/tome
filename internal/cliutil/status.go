package cliutil

import "github.com/kpiljoong/tome/internal/logx"

type SyncStatus struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Filename  string `json:"filename"`
	Source    string `json:"source"` // local, remote, synced, conflict
}

func NewStatus(ns, id, file, source string) SyncStatus {
	return SyncStatus{
		ID:        id,
		Namespace: ns,
		Filename:  file,
		Source:    source,
	}
}

func PrintStatus(statuses []SyncStatus, jsonOut bool) error {
	if jsonOut {
		return PrintPrettyJSON(statuses)
	}

	if len(statuses) == 0 {
		logx.Success("✅ Everything is in sync")
		return nil
	}

	for _, s := range statuses {
		switch s.Source {
		case "local":
			logx.Info("🆕 Only in local:  %s/%s (%s)", s.Namespace, s.Filename, s.ID)
		case "remote":
			logx.Info("☁️ Only in remote: %s/%s (%s)", s.Namespace, s.Filename, s.ID)
		case "synced":
			logx.Info("🔗 Synced:   %s/%s (%s)", s.Namespace, s.Filename, s.ID)
		case "conflict":
			logx.Warn("⚠️ Conflict: %s/%s (%s)", s.Namespace, s.Filename, s.ID)
		}
	}
	return nil
}
