package pipeline

import (
	"regexp"
	"strings"
)

// TitleExtractRule is a per-source regex applied to full HTML before generic <title>.
type TitleExtractRule struct {
	Regex string `json:"regex"`
}

// TitleExtractOpts configures source-specific title extraction and blocklist.
type TitleExtractOpts struct {
	Rules     []TitleExtractRule
	Blocklist []string
}

var globalGenericSiteTitles = map[string]struct{}{
	"奇安信威胁情报中心":                      {},
	"qianxin threat intelligence center": {},
}

// TitleOptsForURL returns collect-source title rules only (no host-specific defaults).
func TitleOptsForURL(_ string, fromCollect *TitleExtractOpts) *TitleExtractOpts {
	return fromCollect
}

func (o *TitleExtractOpts) blocklist() []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(s string) {
		s = NormalizeTitle(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	for k := range globalGenericSiteTitles {
		add(k)
	}
	if o != nil {
		for _, b := range o.Blocklist {
			add(b)
		}
	}
	return out
}

func isBlockedTitle(s string, blocklist []string) bool {
	s = NormalizeTitle(s)
	if s == "" {
		return true
	}
	lower := strings.ToLower(s)
	for _, b := range blocklist {
		if strings.ToLower(b) == lower {
			return true
		}
	}
	if _, ok := globalGenericSiteTitles[lower]; ok {
		return true
	}
	return false
}

func isPlausibleWithBlocklist(s string, blocklist []string) bool {
	return IsPlausibleTitle(s) && !isBlockedTitle(s, blocklist)
}

// IsMergeableIntelTitle reports whether two intel rows may share one attack_knowledge by title match.
func IsMergeableIntelTitle(title string) bool {
	return isPlausibleWithBlocklist(NormalizeTitle(title), (&TitleExtractOpts{}).blocklist())
}

// ExtractTitleFromRules runs configured regexes (first plausible match wins).
func ExtractTitleFromRules(html string, rules []TitleExtractRule, blocklist []string) string {
	for _, rule := range rules {
		re, err := regexp.Compile(rule.Regex)
		if err != nil {
			continue
		}
		m := re.FindStringSubmatch(html)
		if len(m) < 2 {
			continue
		}
		t := cleanTitle(m[1])
		if isPlausibleWithBlocklist(t, blocklist) {
			return t
		}
	}
	return ""
}
