package utils

import (
	"regexp"
	"strconv"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,199}$`)
var chapterKeyPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
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

func ValidChapterKey(value string) bool {
	value = strings.TrimSpace(value)
	if !chapterKeyPattern.MatchString(value) {
		return false
	}
	return value != "0"
}

func NormalizeChapterKey(value string) string {
	value = strings.TrimSpace(value)
	if !strings.Contains(value, ".") {
		return value
	}
	parts := strings.SplitN(value, ".", 2)
	major := strings.TrimLeft(parts[0], "0")
	minor := strings.TrimRight(parts[1], "0")
	if major == "" {
		major = "0"
	}
	if minor == "" {
		return major
	}
	return major + "." + minor
}

func ChapterStorageIndex(value string) int {
	value = NormalizeChapterKey(value)
	if !strings.Contains(value, ".") {
		n, _ := strconv.Atoi(value)
		return n
	}

	parts := strings.SplitN(value, ".", 2)
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	return major*1000 + minor
}
