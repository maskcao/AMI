package pipeline

import (
	"regexp"
	"strings"
)

var (
	blockTags = regexp.MustCompile(`(?is)</(h[1-6]|p|div|li|tr|pre|blockquote)>`)
	h1Re = regexp.MustCompile(`(?is)<h1[^>]*>([\s\S]*?)</h1>`)
	h2Re = regexp.MustCompile(`(?is)<h2[^>]*>([\s\S]*?)</h2>`)
	h3Re = regexp.MustCompile(`(?is)<h3[^>]*>([\s\S]*?)</h3>`)
	h4Re = regexp.MustCompile(`(?is)<h4[^>]*>([\s\S]*?)</h4>`)
	aTagRe    = regexp.MustCompile(`(?is)<a\b[^>]*\bhref=["']([^"']*)["'][^>]*>([\s\S]*?)</a>`)
	imgMdRe   = regexp.MustCompile(`(?is)<img\b[^>]*\balt=["']([^"']*)["'][^>]*\bsrc=["']([^"']+)["'][^>]*>`)
	imgMdRe2  = regexp.MustCompile(`(?is)<img\b[^>]*\bsrc=["']([^"']+)["'][^>]*>`)
	liRe      = regexp.MustCompile(`(?is)<li[^>]*>([\s\S]*?)</li>`)
	brRe      = regexp.MustCompile(`(?is)<br\s*/?>`)
)

// HTMLToMarkdown converts article HTML to readable Markdown (best-effort, no external deps).
func HTMLToMarkdown(html string) string {
	s := html
	s = scriptBlockRe.ReplaceAllString(s, "")
	s = styleBlockRe.ReplaceAllString(s, "")
	s = jsonLDBlockRe.ReplaceAllString(s, "")
	s = schemaBlobRe.ReplaceAllString(s, "")
	s = brRe.ReplaceAllString(s, "\n")

	for level, re := range map[int]*regexp.Regexp{1: h1Re, 2: h2Re, 3: h3Re, 4: h4Re} {
		s = re.ReplaceAllStringFunc(s, func(m string) string {
			sub := re.FindStringSubmatch(m)
			if len(sub) < 2 {
				return m
			}
			text := strings.TrimSpace(stripTags(sub[1]))
			if text == "" {
				return ""
			}
			return "\n\n" + strings.Repeat("#", level) + " " + text + "\n\n"
		})
	}

	s = aTagRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := aTagRe.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		href := strings.TrimSpace(sub[1])
		text := strings.TrimSpace(stripTags(sub[2]))
		if text == "" {
			text = href
		}
		if href == "" {
			return text
		}
		return "[" + text + "](" + href + ")"
	})

	s = imgMdRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := imgMdRe.FindStringSubmatch(m)
		if len(sub) >= 3 {
			alt := strings.TrimSpace(sub[1])
			src := strings.TrimSpace(sub[2])
			if alt == "" {
				alt = "image"
			}
			return "![" + alt + "](" + src + ")"
		}
		return m
	})
	s = imgMdRe2.ReplaceAllStringFunc(s, func(m string) string {
		sub := imgMdRe2.FindStringSubmatch(m)
		if len(sub) >= 2 {
			return "![](" + strings.TrimSpace(sub[1]) + ")"
		}
		return m
	})

	s = liRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := liRe.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		text := strings.TrimSpace(stripTags(sub[1]))
		if text == "" {
			return ""
		}
		return "\n- " + text
	})

	s = stripTags(s)
	s = blockTags.ReplaceAllString(s, "\n\n")
	lines := strings.Split(s, "\n")
	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n\n"))
}

func stripTags(s string) string {
	return strings.TrimSpace(tagRe.ReplaceAllString(s, " "))
}
