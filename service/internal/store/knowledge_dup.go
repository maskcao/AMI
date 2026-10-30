package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DupCandidate struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

const dupScanWindow = 800

func normalizeDupTokens(s string) map[string]struct{} {
	tokens := map[string]struct{}{}
	cur := strings.Builder{}
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			tokens[cur.String()] = struct{}{}
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		tokens[cur.String()] = struct{}{}
	}
	return tokens
}

// tokenSim returns the better of Jaccard and containment so that a short title
// fully contained in a long one still scores high.
func tokenSim(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for t := range a {
		if _, ok := b[t]; ok {
			inter++
		}
	}
	jaccard := float64(inter) / float64(len(a)+len(b)-inter)
	containment := float64(inter) / float64(min(len(a), len(b)))
	return max(jaccard, containment)
}

func nameSet(items []map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, it := range items {
		name, _ := it["name"].(string)
		n := strings.ToLower(strings.TrimSpace(name))
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}

func cveSet(items []map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, it := range items {
		id, _ := it["cveId"].(string)
		n := strings.ToUpper(strings.TrimSpace(id))
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}

func setOverlapRatio(a, b map[string]struct{}) (float64, int) {
	if len(a) == 0 || len(b) == 0 {
		return 0, 0
	}
	inter := 0
	for t := range a {
		if _, ok := b[t]; ok {
			inter++
		}
	}
	return float64(inter) / float64(min(len(a), len(b))), inter
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	for t := range a {
		if _, ok := b[t]; ok {
			inter++
		}
	}
	return float64(inter) / float64(len(a)+len(b)-inter)
}

// findDupCandidates compares one knowledge entry against the most recently
// updated rows and returns scored suspects (score >= 0.45, top 5).
func (d *DB) findDupCandidates(excludeID int64, title string, payload KnowledgePayload) []DupCandidate {
	rows, err := d.Query(`SELECT id, title, payload_json FROM attack_knowledge WHERE id != ? ORDER BY updated_at DESC LIMIT ?`, excludeID, dupScanWindow)
	if err != nil {
		return nil
	}
	defer rows.Close()

	cveA := cveSet(payload.Vulnerabilities)
	techA := map[string]struct{}{}
	for _, t := range payload.Techniques {
		id, _ := t["techniqueId"].(string)
		n := strings.ToUpper(strings.TrimSpace(id))
		if n != "" {
			techA[n] = struct{}{}
		}
	}
	malA := nameSet(payload.Malware)
	titleA := normalizeDupTokens(title)

	var out []DupCandidate
	for rows.Next() {
		var id int64
		var bTitle, bPayload string
		if err := rows.Scan(&id, &bTitle, &bPayload); err != nil {
			continue
		}
		pl, err := ParsePayload(bPayload)
		if err != nil {
			continue
		}
		titleSim := tokenSim(titleA, normalizeDupTokens(bTitle))
		cveRatio, cveInter := setOverlapRatio(cveA, cveSet(pl.Vulnerabilities))
		techSim := jaccard(techA, techSet(pl.Techniques))
		malRatio, malInter := setOverlapRatio(malA, nameSet(pl.Malware))

		score := 0.5*cveRatio + 0.2*titleSim + 0.15*techSim + 0.15*malRatio
		if score < 0.45 {
			continue
		}
		var reasons []string
		if cveInter > 0 {
			reasons = append(reasons, fmt.Sprintf("CVE 重叠 %d 个", cveInter))
		}
		if titleSim >= 0.5 {
			reasons = append(reasons, fmt.Sprintf("标题相似度 %.0f%%", titleSim*100))
		}
		if techSim >= 0.5 {
			reasons = append(reasons, fmt.Sprintf("ATT&CK 手法重叠 %.0f%%", techSim*100))
		}
		if malInter > 0 {
			reasons = append(reasons, fmt.Sprintf("恶意软件重叠 %d 个", malInter))
		}
		out = append(out, DupCandidate{ID: id, Title: bTitle, Score: round2(score), Reasons: reasons})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func techSet(items []map[string]any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, t := range items {
		id, _ := t["techniqueId"].(string)
		n := strings.ToUpper(strings.TrimSpace(id))
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// BackfillDupCandidates recomputes dup candidates for every knowledge row
// (full scan, not just the recent window) and stores the results.
func (d *DB) BackfillDupCandidates() (int, error) {
	rows, err := d.Query(`SELECT id, title, payload_json FROM attack_knowledge ORDER BY id`)
	if err != nil {
		return 0, err
	}
	type entry struct {
		id    int64
		title string
		cve   map[string]struct{}
		tech  map[string]struct{}
		mal   map[string]struct{}
		toks  map[string]struct{}
	}
	var entries []entry
	for rows.Next() {
		var id int64
		var title, pj string
		if err := rows.Scan(&id, &title, &pj); err != nil {
			continue
		}
		pl, err := ParsePayload(pj)
		if err != nil {
			continue
		}
		tech := map[string]struct{}{}
		for _, t := range pl.Techniques {
			n, _ := t["techniqueId"].(string)
			n = strings.ToUpper(strings.TrimSpace(n))
			if n != "" {
				tech[n] = struct{}{}
			}
		}
		entries = append(entries, entry{
			id: id, title: title,
			cve:  cveSet(pl.Vulnerabilities),
			tech: tech,
			mal:  nameSet(pl.Malware),
			toks: normalizeDupTokens(title),
		})
	}
	rows.Close()

	updated := 0
	for i := range entries {
		a := &entries[i]
		var out []DupCandidate
		for j := range entries {
			if i == j {
				continue
			}
			b := &entries[j]
			titleSim := tokenSim(a.toks, b.toks)
			cveRatio, cveInter := setOverlapRatio(a.cve, b.cve)
			techSim := jaccard(a.tech, b.tech)
			malRatio, malInter := setOverlapRatio(a.mal, b.mal)
			score := 0.5*cveRatio + 0.2*titleSim + 0.15*techSim + 0.15*malRatio
			if score < 0.45 {
				continue
			}
			var reasons []string
			if cveInter > 0 {
				reasons = append(reasons, fmt.Sprintf("CVE 重叠 %d 个", cveInter))
			}
			if titleSim >= 0.5 {
				reasons = append(reasons, fmt.Sprintf("标题相似度 %.0f%%", titleSim*100))
			}
			if techSim >= 0.5 {
				reasons = append(reasons, fmt.Sprintf("ATT&CK 手法重叠 %.0f%%", techSim*100))
			}
			if malInter > 0 {
				reasons = append(reasons, fmt.Sprintf("恶意软件重叠 %d 个", malInter))
			}
			out = append(out, DupCandidate{ID: b.id, Title: b.title, Score: round2(score), Reasons: reasons})
		}
		sort.Slice(out, func(x, y int) bool { return out[x].Score > out[y].Score })
		if len(out) > 5 {
			out = out[:5]
		}
		blob, _ := json.Marshal(out)
		if _, err := d.Exec(`UPDATE attack_knowledge SET dup_candidates_json=? WHERE id=?`, string(blob), a.id); err != nil {
			return updated, err
		}
		if len(out) > 0 {
			updated++
		}
	}
	return updated, nil
}
