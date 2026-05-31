package llm

import (
	"regexp"
	"strings"
)

var (
	cveFmtRe = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	techFmtRe = regexp.MustCompile(`^T\d{4}(?:\.\d{3})?$`)
)

// QualityReport captures post-enrichment checks for UI and ops logs.
type QualityReport struct {
	Hints []string `json:"hints"`
	OK    bool     `json:"ok"`
}

func ValidateEnrich(er *EnrichResult) QualityReport {
	if er == nil {
		return QualityReport{OK: false, Hints: []string{"enrich: empty result"}}
	}
	var hints []string
	ok := true

	sum := strings.TrimSpace(er.Summary)
	if sum == "" {
		hints = append(hints, "overview: missing")
		ok = false
	} else if len([]rune(sum)) < 20 {
		hints = append(hints, "overview: too short")
		ok = false
	} else if strings.Contains(sum, `"@context"`) || strings.Contains(sum, "schema.org") {
		hints = append(hints, "overview: contains structured metadata noise")
		ok = false
	} else if len([]rune(sum)) > 600 {
		hints = append(hints, "overview: unusually long")
	}

	for _, c := range er.SuggestedCves {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if !cveFmtRe.MatchString(strings.ToUpper(c)) {
			hints = append(hints, "cve: invalid format "+c)
			ok = false
		}
	}
	for _, t := range er.SuggestedTechniques {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if !techFmtRe.MatchString(strings.ToUpper(t)) {
			hints = append(hints, "technique: invalid id "+t)
			ok = false
		}
	}
	if len(er.SuggestedMalware) == 0 {
		hints = append(hints, "malware: none extracted")
	}
	if len(er.SuggestedTechniques) == 0 {
		hints = append(hints, "techniques: none extracted")
	}
	if len(er.Tools) == 0 {
		hints = append(hints, "tools: none extracted")
	}
	if len(er.Infrastructure) == 0 {
		hints = append(hints, "infrastructure: none extracted")
	}
	for _, m := range er.Malware {
		if strings.TrimSpace(m.Name) == "" {
			hints = append(hints, "malware: empty name entry")
			ok = false
		}
		if strings.TrimSpace(m.Classification) == "" {
			hints = append(hints, "malware: missing classification for "+m.Name)
		}
	}
	if len(hints) == 0 {
		hints = []string{"quality: pass"}
	}
	return QualityReport{Hints: hints, OK: ok}
}
