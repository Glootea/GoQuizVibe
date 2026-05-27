package services

import (
	"regexp"
	"strings"
)

var (
	cacheUUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func normalizeCacheKey(key string) string {
	segments := strings.Split(key, ":")
	if len(segments) == 0 {
		return key
	}

	last := segments[len(segments)-1]
	if cacheUUIDRegex.MatchString(last) {
		segments[len(segments)-1] = "{id}"
	}

	normalized := strings.Join(segments, ":")
	normalized = strings.ReplaceAll(normalized, ":{id}", "")
	normalized = strings.ReplaceAll(normalized, "{id}", "")

	return normalized
}