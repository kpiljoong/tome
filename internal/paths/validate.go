package paths

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateNamespace(namespace string) error {
	return validatePathSegment("namespace", namespace)
}

func ValidateEntryID(id string) error {
	if err := validatePathSegment("entry ID", id); err != nil {
		return err
	}
	if strings.HasSuffix(id, ".json") {
		return fmt.Errorf("entry ID must not include a .json suffix")
	}
	return nil
}

func ValidateBlobHash(hash string) error {
	if hash == "" {
		return fmt.Errorf("blob hash must not be empty")
	}
	return validatePathSegment("blob hash", SanitizeHash(hash))
}

func SafeBlobPath(hash string) (string, error) {
	if err := ValidateBlobHash(hash); err != nil {
		return "", err
	}
	return BlobPath(hash), nil
}

func SafeNamespaceDir(namespace string) (string, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}
	return NamespaceDir(namespace), nil
}

func SafeJournalPath(namespace, id string) (string, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}
	if err := ValidateEntryID(id); err != nil {
		return "", err
	}
	return JournalPath(namespace, id), nil
}

func validatePathSegment(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	if value == "." || value == ".." || filepath.IsAbs(value) || filepath.Clean(value) != value || strings.ContainsAny(value, "/\\") {
		return fmt.Errorf("invalid %s path segment %q", name, value)
	}
	return nil
}
