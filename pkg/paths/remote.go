package paths

import "path/filepath"

func RemoteJournalPath(namespace, id string) string {
	return filepath.Join(RemoteJournalsPrefix, namespace, id+".json")
}

func RemoteBlobPath(hash string) string {
	return filepath.Join(RemoteBlobsPrefix, hash)
}
