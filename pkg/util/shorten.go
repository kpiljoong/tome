package util

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func ShortenURL(longURL string) (string, error) {
	endpoint := "https://tinyurl.com/api-create.php?url=" + url.QueryEscape(longURL)
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to contact TinyURL: %w", err)
	}
	defer resp.Body.Close()

	short, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	return string(short), nil
}
