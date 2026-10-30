package ingest

import (
	"html"
	"regexp"
	"strings"
)

// WechatWeeklyEntry is one intel block inside a Qianxin weekly digest article.
type WechatWeeklyEntry struct {
	Category    string // attack_group | malware | other
	Title       string
	DisclosedAt string
	SourceURL   string
	RelatedInfo string
}

var (
	weeklyDisclosedRe = regexp.MustCompile(`(?is)披露时间[：:](?:\s*<[^>]+>)*\s*([^<]+)`)
	weeklySourceRe    = regexp.MustCompile(`(?is)情报来源[：:](?:\s*<[^>]+>)*\s*([^\s<]+)`)
	weeklySectionHdr  = regexp.MustCompile(`(?is)(攻击团伙情报|恶意代码情报|攻击行动或事件情报|漏洞相关情报)`)
)

// ParseWechatWeeklyThreatDigest extracts structured rows from 奇安信 weekly digest HTML.
func ParseWechatWeeklyThreatDigest(rawHTML []byte) []WechatWeeklyEntry {
	s := string(rawHTML)
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, `\x3c`, "<")
	s = strings.ReplaceAll(s, `\x3e`, ">")
	s = strings.ReplaceAll(s, `\x22`, `"`)

	headers := weeklySectionHdr.FindAllStringIndex(s, -1)
	if len(headers) == 0 {
		return parseWeeklyBlocks(s, "other")
	}

	var out []WechatWeeklyEntry
	for i, loc := range headers {
		cat := categoryFromHeader(weeklySectionHdr.FindString(s[loc[0]:loc[1]]))
		end := len(s)
		if i+1 < len(headers) {
			end = headers[i+1][0]
		}
		out = append(out, parseWeeklyBlocks(s[loc[1]:end], cat)...)
	}
	return out
}

func categoryFromHeader(h string) string {
	switch {
	case strings.Contains(h, "攻击团伙"):
		return "attack_group"
	case strings.Contains(h, "恶意代码"):
		return "malware"
	default:
		return "other"
	}
}

func parseWeeklyBlocks(fragment, category string) []WechatWeeklyEntry {
	parts := regexp.MustCompile(`(?is)披露时间[：:]`).Split(fragment, -1)
	if len(parts) <= 1 {
		return nil
	}
	var out []WechatWeeklyEntry
	for i := 1; i < len(parts); i++ {
		block := "披露时间:" + parts[i]
		if j := strings.Index(block, "披露时间:"); j > 0 {
			block = block[:j]
		}
		e := WechatWeeklyEntry{Category: category}
		if m := weeklyDisclosedRe.FindStringSubmatch(block); len(m) > 1 {
			e.DisclosedAt = strings.TrimSpace(stripTags(m[1]))
		}
		if m := weeklySourceRe.FindStringSubmatch(block); len(m) > 1 {
			e.SourceURL = strings.TrimSpace(stripTags(m[1]))
		}
		if idx := strings.Index(block, "相关信息"); idx >= 0 {
			rest := block[idx:]
			if j := strings.Index(rest, "</p>"); j > 0 {
				rest = rest[j+4:]
			}
			e.RelatedInfo = trimToLen(stripTags(rest), 2048)
		}
		pre := parts[i-1]
		if len(pre) > 500 {
			pre = pre[len(pre)-500:]
		}
		e.Title = guessWeeklyTitle(pre, e.RelatedInfo)
		if e.SourceURL == "" && e.DisclosedAt == "" && e.RelatedInfo == "" {
			continue
		}
		out = append(out, e)
	}
	return out
}

func guessWeeklyTitle(before, related string) string {
	// Prefer first list item text in section preamble
	liRe := regexp.MustCompile(`(?is)<li[^>]*>[\s\S]*?<span[^>]*>([^<]+)</span>`)
	if m := liRe.FindStringSubmatch(before); len(m) > 1 {
		t := strings.TrimSpace(stripTags(m[1]))
		if t != "" {
			return t
		}
	}
	if related != "" {
		r := []rune(related)
		if len(r) > 80 {
			return string(r[:80]) + "…"
		}
		return related
	}
	return ""
}

func stripTags(s string) string {
	s = regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(s, " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

func trimToLen(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ExpandWeeklyDigestSourceLinks returns child article URLs from weekly digests (attack_group + malware).
func ExpandWeeklyDigestSourceLinks(cfg SourceConfig, queue []listLink) ([]listLink, error) {
	if cfg.DigestParser != "wechat_weekly_threat" {
		return queue, nil
	}
	seen := map[string]struct{}{}
	var out []listLink
	add := func(lk listLink) {
		if lk.URL == "" {
			return
		}
		if _, ok := seen[lk.URL]; ok {
			return
		}
		seen[lk.URL] = struct{}{}
		out = append(out, lk)
	}
	limit := cfg.DigestSourceMax
	unlimited := limit <= 0
	added := 0
	for _, wk := range queue {
		if !unlimited && added >= limit {
			break
		}
		body, err := fetchArticleHTML(wk.URL, cfg.TimeoutSec)
		if err != nil {
			continue
		}
		for _, e := range ParseWechatWeeklyThreatDigest(body) {
			if e.Category != "attack_group" && e.Category != "malware" {
				continue
			}
			if e.SourceURL == "" {
				continue
			}
			title := e.Title
			if title == "" {
				title = e.SourceURL
			}
			add(listLink{URL: e.SourceURL, Title: title})
			added++
			if !unlimited && added >= limit {
				break
			}
		}
		sleepDelay(cfg.DelayMS)
		if cfg.ImportWeeklyDigest {
			add(wk)
		}
	}
	return out, nil
}
