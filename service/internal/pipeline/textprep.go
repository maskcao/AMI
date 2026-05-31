package pipeline

import (
	"regexp"
	"strings"
)

var (
	scriptBlockRe = regexp.MustCompile(`(?is)<script[^>]*>[\s\S]*?</script>`)
	styleBlockRe  = regexp.MustCompile(`(?is)<style[^>]*>[\s\S]*?</style>`)
	noscriptRe    = regexp.MustCompile(`(?is)<noscript[^>]*>[\s\S]*?</noscript>`)
	jsonLDBlockRe = regexp.MustCompile(`(?is)<script[^>]+type=["']application/ld\+json["'][^>]*>[\s\S]*?</script>`)
	schemaBlobRe  = regexp.MustCompile(`(?is)\{\s*"@context"\s*:\s*"https?://schema\.org"[\s\S]*?\}\s*`)
	htmlCommentRe = regexp.MustCompile(`(?is)<!--[\s\S]*?-->`)
	ghostRSSRe    = regexp.MustCompile(`(?is)<rss[\s\S]*?</rss>`)
)

const maxEnrichChars = 10000

// PrepareArticleText strips boilerplate before rules/LLM enrichment (stateless per article).
func PrepareArticleText(html string) string {
	s := html
	s = scriptBlockRe.ReplaceAllString(s, " ")
	s = styleBlockRe.ReplaceAllString(s, " ")
	s = noscriptRe.ReplaceAllString(s, " ")
	s = jsonLDBlockRe.ReplaceAllString(s, " ")
	s = htmlCommentRe.ReplaceAllString(s, " ")
	s = ghostRSSRe.ReplaceAllString(s, " ")
	s = schemaBlobRe.ReplaceAllString(s, " ")
	text := HTMLToText(s)
	text = collapseNoiseLines(text)
	if len(text) > maxEnrichChars {
		text = text[:maxEnrichChars]
	}
	return strings.TrimSpace(text)
}

func collapseNoiseLines(text string) string {
	lines := strings.Split(text, "\n")
	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 4 && strings.Count(line, "http") >= 3 && len(line) < 400 {
			continue
		}
		if strings.HasPrefix(line, "{") && strings.Contains(line, `"@type"`) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
