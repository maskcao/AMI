package pipeline

import (
	"regexp"
	"strings"
)

var snapshotImgTagRe = regexp.MustCompile(`(?is)<img\b([^>]*)>`)

// HydrateLazyLoadImages copies data-src (and data-srcset) to src/srcset when src is missing.
// Unit42 and similar sites use lozad; archived HTML keeps data-src only and images break in SPA v-html.
func HydrateLazyLoadImages(html string) string {
	return snapshotImgTagRe.ReplaceAllStringFunc(html, func(full string) string {
		sub := snapshotImgTagRe.FindStringSubmatch(full)
		if len(sub) < 2 {
			return full
		}
		attrs := sub[1]
		dataSrc := htmlAttrValue(attrs, "data-src")
		if dataSrc == "" {
			return full
		}
		if src := htmlAttrValue(attrs, "src"); src != "" {
			return full
		}
		prefix := `src="` + escapeHTMLAttr(dataSrc) + `" `
		if ds := htmlAttrValue(attrs, "data-srcset"); ds != "" && htmlAttrValue(attrs, "srcset") == "" {
			prefix += `srcset="` + escapeHTMLAttr(ds) + `" `
		}
		return "<img " + prefix + strings.TrimSpace(attrs) + ">"
	})
}

var htmlAttrValueRe = regexp.MustCompile(`(?is)\b([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)')`)

func htmlAttrValue(attrs, name string) string {
	needle := strings.ToLower(name)
	for _, m := range htmlAttrValueRe.FindAllStringSubmatch(attrs, -1) {
		if len(m) < 4 {
			continue
		}
		if strings.ToLower(m[1]) != needle {
			continue
		}
		if m[3] != "" {
			return m[3]
		}
		return m[4]
	}
	return ""
}

func escapeHTMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	return strings.ReplaceAll(s, `"`, "&quot;")
}
