package paths

import "path/filepath"

func RemoteJournalPath(namespace, id string) string {
	return filepath.Join(RemoteJournalsPrefix, namespace, id+".json")
}

func RemoteBlobPath(hash string) string {
	return filepath.ToSlash(filepath.Join(RemoteBlobsPrefix, sanitizeHash(hash)))
}

func RemoteNamespacePrefix(namespace string) string {
	return filepath.ToSlash(filepath.Join(RemoteJournalsPrefix, namespace))
}
