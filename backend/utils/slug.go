package utils

import (
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,199}$`)
var slugReplacePattern = regexp.MustCompile(`[^a-z0-9]+`)

func ValidSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

func Slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = slugReplacePattern.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "unknown"
	}
	return slug
}
