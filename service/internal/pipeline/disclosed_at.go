package pipeline

import (
	"regexp"
	"strings"
	"time"
)

var (
	metaPublishedRe = regexp.MustCompile(`(?is)<meta\b[^>]*(?:property|name)=["'](?:article:published_time|og:article:published_time|publish(?:ed)?[_-]?date|date)["'][^>]*\bcontent=["']([^"']+)["']`)
	metaPublishedRe2 = regexp.MustCompile(`(?is)<meta\b[^>]*\bcontent=["']([^"']+)["'][^>]*(?:property|name)=["'](?:article:published_time|og:article:published_time|publish(?:ed)?[_-]?date|date)["']`)
	jsonDatePublishedRe = regexp.MustCompile(`(?is)"datePublished"\s*:\s*"([^"]+)"`)
	nuxtPublishTimeRe = regexp.MustCompile(`(?is)publish_time\s*=\s*["']([^"']+)["']`)
	timeDatetimeRe = regexp.MustCompile(`(?is)<time\b[^>]*\bdatetime=["']([^"']+)["']`)
)

// ExtractDisclosedAtFromHTML returns RFC3339 UTC for article publish/event time when found.
func ExtractDisclosedAtFromHTML(html string) string {
	html = strings.TrimSpace(html)
	if html == "" {
		return ""
	}
	for _, re := range []*regexp.Regexp{
		metaPublishedRe, metaPublishedRe2,
		jsonDatePublishedRe,
		nuxtPublishTimeRe,
		timeDatetimeRe,
	} {
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			if ts := normalizeDisclosedTime(m[1]); ts != "" {
				return ts
			}
		}
	}
	return ""
}

func normalizeDisclosedTime(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	if len(raw) >= 10 {
		if t, err := time.Parse("2006-01-02", raw[:10]); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}
	return ""
}
