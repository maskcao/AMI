package pipeline

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxTitleRunes = 180

var tagRe = regexp.MustCompile(`<[^>]+>`)

var (
	ogTitleContentFirst = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	ogTitlePropertyLast = regexp.MustCompile(`(?is)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:title["']`)
	titleTagRe          = regexp.MustCompile(`(?is)<title[^>]*>([\s\S]*?)</title>`)
	h1TagRe             = regexp.MustCompile(`(?is)<h1[^>]*>([\s\S]*?)</h1>`)
	ghostPostTitleH1    = regexp.MustCompile(`(?is)<h1[^>]*class="[^"]*\bpost-full-title\b[^"]*"[^>]*>([\s\S]*?)</h1>`)
	jsonHeadlineRe      = regexp.MustCompile(`(?is)"headline"\s*:\s*"((?:\\.|[^"\\])*)"`)
)

func HTMLToText(html string) string {
	s := tagRe.ReplaceAllString(html, " ")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func Snippet(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "…"
}

var cveRe = regexp.MustCompile(`CVE-\d{4}-\d+`)

// ExtractTitleFromHTML picks article title from structured fields, not list-card anchor text.
func ExtractTitleFromHTML(html string) string {
	for _, re := range []*regexp.Regexp{ogTitleContentFirst, ogTitlePropertyLast} {
		if m := re.FindStringSubmatch(html); len(m) > 1 {
			if t := cleanTitle(m[1]); IsPlausibleTitle(t) {
				return t
			}
		}
	}
	if m := jsonHeadlineRe.FindStringSubmatch(html); len(m) > 1 {
		if t := cleanTitle(unescapeJSONString(m[1])); IsPlausibleTitle(t) {
			return t
		}
	}
	if m := ghostPostTitleH1.FindStringSubmatch(html); len(m) > 1 {
		if t := cleanTitle(m[1]); IsPlausibleTitle(t) {
			return t
		}
	}
	if m := titleTagRe.FindStringSubmatch(html); len(m) > 1 {
		if t := cleanTitle(m[1]); IsPlausibleTitle(t) {
			return t
		}
	}
	if m := h1TagRe.FindStringSubmatch(html); len(m) > 1 {
		if t := cleanTitle(m[1]); IsPlausibleTitle(t) {
			return t
		}
	}
	return ""
}

func unescapeJSONString(s string) string {
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

func cleanTitle(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimSpace(s)
	for _, sep := range []string{" - ", " | ", " – ", " — ", " · "} {
		if i := strings.Index(s, sep); i > 8 {
			left := strings.TrimSpace(s[:i])
			if IsPlausibleTitle(left) || utf8.RuneCountInString(left) <= maxTitleRunes {
				s = left
				break
			}
		}
	}
	return NormalizeTitle(s)
}

// NormalizeTitle trims and caps length for DB/UI display.
func NormalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if i := strings.IndexAny(s, "\n\r\t"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if utf8.RuneCountInString(s) > maxTitleRunes {
		r := []rune(s)
		s = string(r[:maxTitleRunes]) + "…"
	}
	return s
}

// IsPlausibleTitle rejects list-card blurbs and body text mistaken as titles.
func IsPlausibleTitle(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	n := utf8.RuneCountInString(s)
	if n < 4 || n > maxTitleRunes {
		return false
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return false
	}
	if strings.Count(s, "。") >= 2 || strings.Count(s, "！") >= 2 {
		return false
	}
	if strings.Count(s, ". ") >= 3 {
		return false
	}
	for _, kw := range []string{"版权", "版权声明", "转载", "背景介绍", "概述", "点击阅读原文"} {
		if strings.Contains(s, kw) && n > 60 {
			return false
		}
	}
	return true
}

// TitleFromURL derives a short label from the last path segment.
func TitleFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}
	seg := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		seg = path[i+1:]
	}
	seg = strings.ReplaceAll(seg, "-", " ")
	seg = strings.ReplaceAll(seg, "_", " ")
	return NormalizeTitle(seg)
}

// PickIngestTitle prefers HTML title over list-page anchor text.
func PickIngestTitle(html []byte, pageURL, listHint string) string {
	if t := ExtractTitleFromHTML(string(html)); t != "" {
		return t
	}
	hint := NormalizeTitle(listHint)
	if IsPlausibleTitle(hint) {
		return hint
	}
	if t := TitleFromURL(pageURL); t != "" {
		return t
	}
	return NormalizeTitle(Snippet(hint, 80))
}

// ResolveIntelTitle fixes bad titles before knowledge write.
func ResolveIntelTitle(current, sourceURL string, html []byte) string {
	cur := NormalizeTitle(current)
	if IsPlausibleTitle(cur) {
		return cur
	}
	if len(html) > 0 {
		if t := ExtractTitleFromHTML(string(html)); t != "" {
			return t
		}
	}
	if t := TitleFromURL(sourceURL); t != "" {
		return t
	}
	if cur != "" {
		return NormalizeTitle(Snippet(cur, 80))
	}
	return TitleFromURL(sourceURL)
}

// ShouldRefreshIntelTitle returns true when HTML title should replace ingest placeholder.
func ShouldRefreshIntelTitle(current, sourceURL, fromHTML string) bool {
	fromHTML = NormalizeTitle(fromHTML)
	if !IsPlausibleTitle(fromHTML) {
		return false
	}
	cur := NormalizeTitle(current)
	if !IsPlausibleTitle(cur) {
		return true
	}
	if sourceURL != "" && cur == sourceURL {
		return true
	}
	if strings.HasPrefix(cur, "http://") || strings.HasPrefix(cur, "https://") {
		return true
	}
	if utf8.RuneCountInString(cur) > maxTitleRunes {
		return true
	}
	return false
}

func ExtractCVEs(text string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range cveRe.FindAllString(text, -1) {
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}
