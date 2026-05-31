package ingest

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/ami/ami/service/internal/pipeline"
)

type listLink struct {
	URL   string
	Title string
}

func extractArticleLinks(html []byte, baseURL string, cfg SourceConfig) ([]listLink, error) {
	re := regexp.MustCompile(`(?i)<a\s+[^>]*href\s*=\s*["']([^"'#]+)["'][^>]*>([\s\S]*?)</a>`)
	matches := re.FindAllStringSubmatch(string(html), -1)
	base, _ := url.Parse(baseURL)
	seen := map[string]struct{}{}
	var out []listLink
	lex := cfg.LinkExtract
	must := lex.MustContain
	if must == "" {
		must = cfg.LinkSelector
	}
	var preferRe *regexp.Regexp
	if lex.PreferPathRegex != "" {
		preferRe = regexp.MustCompile(lex.PreferPathRegex)
	}

	for _, m := range matches {
		href := strings.TrimSpace(m[1])
		if href == "" || strings.HasPrefix(strings.ToLower(href), "javascript:") {
			continue
		}
		abs := resolveURL(base, href)
		if abs == "" {
			continue
		}
		if must != "" && !strings.Contains(abs, must) {
			continue
		}
		if linkExcluded(abs, lex.ExcludeContains) {
			continue
		}
		if preferRe != nil && !preferRe.MatchString(abs) {
			if strings.Contains(abs, "/tag/") || strings.Contains(abs, "/page/") {
				continue
			}
		}
		title := stripHTML(m[2])
		if !pipeline.IsPlausibleTitle(title) {
			title = ""
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		out = append(out, listLink{URL: abs, Title: title})
	}
	return out, nil
}

func linkExcluded(abs string, excludes []string) bool {
	for _, ex := range excludes {
		if ex != "" && strings.Contains(abs, ex) {
			return true
		}
	}
	return false
}

func stripHTML(s string) string {
	s = regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(s, " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
