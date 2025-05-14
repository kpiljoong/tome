package util

import (
	"regexp"
	"strings"
)

var slugifyRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func Slugify(input string) string {
	slug := slugifyRegex.ReplaceAllString(strings.ToLower(input), "-")
	return strings.Trim(slug, "-")
}
