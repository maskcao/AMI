package pipeline

import (
	"regexp"
	"strings"
)

var (
	aptIDRe     = regexp.MustCompile(`(?i)\b((?:APT|UTG)-Q-\d+)\b`)
	parenENRe   = regexp.MustCompile(`(?i)\b([A-Za-z][A-Za-z0-9._-]{1,40})\s*[（(]\s*((?:APT|UTG)-Q-\d+)\s*[）)]`)
	cnParenAPTRe = regexp.MustCompile(`([\p{Han}·]{2,12})[（(]\s*((?:APT|UTG)-Q-\d+)\s*[）)]`)
	cnOrgRe     = regexp.MustCompile(`([\p{Han}]{2,8})组织`)
	enGroupRe   = regexp.MustCompile(`(?i)\b([A-Z][A-Za-z][A-Za-z0-9._-]{0,39})\s+group\b`)
	operationRe = regexp.MustCompile(`(?i)\bOperation\s+([A-Z][A-Za-z0-9_-]{2,32})\b`)
)

var threatActorBlocklist = map[string]struct{}{
	"apt": {}, "apt group": {}, "apt组织": {}, "攻击组织": {}, "威胁组织": {}, "攻击者": {},
	"攻击团伙": {}, "黑客组织": {}, "未知": {}, "unknown": {},
	"360": {}, "netlab": {}, "qianxin": {}, "奇安信": {},
}

// ExtractThreatActors finds named threat groups from article title and body (rules-based).
func ExtractThreatActors(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	add := func(name string) {
		name = NormalizeThreatActorName(name)
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if _, blocked := threatActorBlocklist[key]; blocked {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}

	for _, m := range parenENRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 3 {
			add(formatActorAlias(m[1], m[2]))
		}
	}
	for _, m := range cnParenAPTRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 3 {
			add(formatActorAlias(m[1], m[2]))
		}
	}
	for _, m := range aptIDRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 2 {
			add(strings.ToUpper(m[1]))
		}
	}
	for _, m := range cnOrgRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 2 {
			add(strings.TrimSpace(m[1]) + "组织")
		}
	}
	for _, m := range enGroupRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimSpace(m[1])
		if aptIDRe.MatchString(name) {
			add(strings.ToUpper(name))
			continue
		}
		add(name + " group")
	}
	for _, m := range operationRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 2 {
			add("Operation " + m[1])
		}
	}
	return collapseThreatActors(out)
}

func formatActorAlias(alias, aptID string) string {
	alias = strings.TrimSpace(alias)
	aptID = strings.ToUpper(strings.TrimSpace(aptID))
	if alias == "" {
		return aptID
	}
	if aptID == "" {
		return alias
	}
	return alias + " (" + aptID + ")"
}

// NormalizeThreatActorName trims noise from extracted actor labels.
func NormalizeThreatActorName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "「」\"'[]")
	if name == "" {
		return ""
	}
	if strings.EqualFold(name, "—") || strings.EqualFold(name, "-") {
		return ""
	}
	lower := strings.ToLower(name)
	for _, frag := range []string{"披露", "情报中心", "威胁情报", "南亚apt", "apt组织", "攻击组织"} {
		if strings.Contains(lower, frag) {
			return ""
		}
	}
	for _, frag := range []string{"blog", "netlab", "360"} {
		if strings.Contains(lower, frag) && !strings.Contains(lower, "apt-q") {
			return ""
		}
	}
	if strings.Contains(name, "(APT-") || strings.Contains(name, "（APT-") {
		if !strings.Contains(name, ")") && !strings.Contains(name, "）") {
			return ""
		}
	}
	if len([]rune(name)) > 64 {
		return ""
	}
	return name
}

func collapseThreatActors(out []string) []string {
	aliasForAPT := map[string]struct{}{}
	for _, a := range out {
		ids := aptIDRe.FindAllString(a, -1)
		if len(ids) == 0 {
			continue
		}
		if len(a) > len(ids[0])+4 {
			for _, id := range ids {
				aliasForAPT[strings.ToUpper(id)] = struct{}{}
			}
		}
	}
	var res []string
	seen := map[string]struct{}{}
	for _, a := range out {
		upper := strings.ToUpper(a)
		if aptIDRe.MatchString(a) && len(a) <= 12 {
			if _, ok := aliasForAPT[upper]; ok {
				continue
			}
		}
		if strings.HasSuffix(strings.ToLower(a), " group") {
			base := strings.TrimSuffix(a, " group")
			if _, ok := aliasForAPT[strings.ToUpper(base)]; ok {
				continue
			}
		}
		key := strings.ToLower(a)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		res = append(res, a)
	}
	return res
}

// MergeThreatActorsExported merges existing payload actors with rule-based extraction from text.
func MergeThreatActorsExported(existing []string, title, summary string) []string {
	return mergeThreatActors(existing, ExtractThreatActors(title+"\n"+summary))
}

func mergeThreatActors(parts ...[]string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, list := range parts {
		for _, a := range list {
			a = NormalizeThreatActorName(a)
			if a == "" {
				continue
			}
			key := strings.ToLower(a)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, a)
		}
	}
	return out
}

func appendUniqueStrings(base []string, more ...string) []string {
	seen := map[string]struct{}{}
	for _, s := range base {
		s = NormalizeThreatActorName(s)
		if s == "" {
			continue
		}
		seen[strings.ToLower(s)] = struct{}{}
	}
	for _, s := range more {
		s = NormalizeThreatActorName(s)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		base = append(base, s)
	}
	return base
}
