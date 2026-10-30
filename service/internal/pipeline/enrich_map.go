package pipeline

import (
	"strings"

	"github.com/ami/ami/service/internal/llm"
	"github.com/ami/ami/service/internal/mitre"
	"github.com/ami/ami/service/internal/store"
)

func applyEnrichToPayload(pl store.KnowledgePayload, er *llm.EnrichResult, cves []string) store.KnowledgePayload {
	if er == nil {
		return pl
	}
	for _, tid := range er.SuggestedTechniques {
		tid = strings.TrimSpace(strings.ToUpper(tid))
		if tid == "" {
			continue
		}
		tactic := mitre.TacticForTechnique(tid)
		if tactic == "" {
			tactic = "—"
		}
		pl.Techniques = appendUniqueMap(pl.Techniques, "techniqueId", tid, map[string]any{
			"techniqueId": tid, "name": tid, "tactic": tactic,
		})
	}
	for _, c := range cves {
		c = strings.TrimSpace(strings.ToUpper(c))
		if c == "" {
			continue
		}
		pl.Vulnerabilities = appendUniqueMap(pl.Vulnerabilities, "cveId", c, map[string]any{
			"cveId": c, "description": "extracted from intel", "inWild": true,
		})
	}
	for _, t := range er.Tools {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		pl.Tools = appendUniqueMap(pl.Tools, "name", name, map[string]any{
			"name": name, "role": strings.TrimSpace(t.Role),
		})
	}
	for _, inf := range er.Infrastructure {
		val := strings.TrimSpace(inf.Value)
		if val == "" {
			continue
		}
		typ := strings.TrimSpace(inf.Type)
		if typ == "" {
			typ = "other"
		}
		pl.Infrastructure = appendUniqueMap(pl.Infrastructure, "value", val, map[string]any{
			"type": typ, "value": val,
		})
	}
	for _, m := range er.Malware {
		name := strings.TrimSpace(m.Name)
		if name == "" {
			continue
		}
		cls := strings.TrimSpace(m.Classification)
		if cls == "" {
			cls = "other"
		}
		pl.Malware = appendUniqueMap(pl.Malware, "name", name, map[string]any{
			"name": name, "class": cls, "classification": cls,
		})
	}
	for _, name := range er.SuggestedMalware {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		pl.Malware = appendUniqueMap(pl.Malware, "name", name, map[string]any{
			"name": name, "class": "other", "classification": "other",
		})
	}
	pl.ThreatActors = appendUniqueStrings(pl.ThreatActors, er.ThreatActors...)
	if len(er.Quality.Hints) > 0 {
		pl.QualityHints = appendUnique(pl.QualityHints, er.Quality.Hints...)
	}
	return pl
}

func appendUniqueMap(slice []map[string]any, key, val string, row map[string]any) []map[string]any {
	for _, existing := range slice {
		if s, _ := existing[key].(string); s == val {
			return slice
		}
	}
	return append(slice, row)
}

func aiExtractFromEnrich(er *llm.EnrichResult, fallbackSummary string, cves []string) *store.AIExtract {
	ai := &store.AIExtract{
		Summary:             fallbackSummary,
		SuggestedTechniques: []string{},
		SuggestedMalware:    []string{},
		SuggestedCves:       cves,
	}
	if er == nil {
		return ai
	}
	if er.Summary != "" {
		ai.Summary = er.Summary
	}
	ai.SuggestedTechniques = append([]string{}, er.SuggestedTechniques...)
	ai.SuggestedMalware = append([]string{}, er.SuggestedMalware...)
	ai.SuggestedCves = cves
	for _, t := range er.Tools {
		if t.Name != "" {
			ai.Tools = append(ai.Tools, t.Name)
		}
	}
	for _, inf := range er.Infrastructure {
		if inf.Value != "" {
			ai.Infrastructure = append(ai.Infrastructure, inf.Type+":"+inf.Value)
		}
	}
	classes := map[string]struct{}{}
	for _, m := range er.Malware {
		if m.Classification != "" {
			classes[m.Classification] = struct{}{}
		}
	}
	for c := range classes {
		ai.MalwareClasses = append(ai.MalwareClasses, c)
	}
	ai.QualityHints = er.Quality.Hints
	ai.QualityOK = er.Quality.OK
	return ai
}
