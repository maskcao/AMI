package analytics

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ami/ami/service/internal/malwareclass"
	"github.com/ami/ami/service/internal/mitre"
	"github.com/ami/ami/service/internal/store"
)

type LinkedRef struct {
	KnowledgeID    int64  `json:"knowledgeId"`
	KnowledgeTitle string `json:"knowledgeTitle"`
}

type Bundle struct {
	Dashboard     any `json:"dashboard"`
	Nav           any `json:"nav"`
	Cve           any `json:"cve"`
	Malware       any `json:"malware"`
	Actors        any `json:"actors"`
	ActorProfiles any `json:"actorProfiles"`
	Techniques    any `json:"techniques"`
	Tools         any `json:"tools"`
	Infra         any `json:"infra"`
	Ioc           any `json:"ioc"`
}

type kAgg struct {
	ID          int64
	Title       string
	Matrix      string
	Industries  []string
	UpdatedAt   string
	DisclosedAt string
	Quality     int
	IntelCount  int
	Payload     store.KnowledgePayload
}

func asMapSlice(v any) []map[string]any {
	if s, ok := v.([]map[string]any); ok {
		return s
	}
	return nil
}

func nonNilSlice(out []map[string]any) []map[string]any {
	if out == nil {
		return []map[string]any{}
	}
	return out
}

func BuildBundle(db *store.DB) (*Bundle, error) {
	rows, _, err := db.ListKnowledge(store.KnowledgeListParams{Limit: 5000})
	if err != nil {
		return nil, err
	}
	var knowledges []kAgg
	for _, r := range rows {
		pl, _ := store.ParsePayload(r.PayloadJSON)
		var ind []string
		_ = json.Unmarshal([]byte(r.TargetIndustriesJSON), &ind)
		knowledges = append(knowledges, kAgg{
			ID: r.ID, Title: r.Title, Matrix: r.AttackMatrix, Industries: ind,
			UpdatedAt: r.UpdatedAt, DisclosedAt: r.EarliestDisclosedAt.String,
			Quality: r.QualityScore, IntelCount: r.IntelCount, Payload: pl,
		})
	}
	intelRows, _, err := db.ListIntel(store.IntelListParams{Limit: 5000})
	if err != nil {
		return nil, err
	}

	cves := buildCves(knowledges)
	malware := buildMalware(knowledges)
	actors, profiles := buildActors(knowledges)
	techniques := buildTechniques(knowledges)
	tools := buildTools(knowledges)
	infra := buildInfra(knowledges)
	iocs := buildIocs(db, intelRows, knowledges)

	snap := map[string]any{
		"knowledgeCount":   len(knowledges),
		"cveCount":         len(asMapSlice(cves)),
		"malwareCount":     len(asMapSlice(malware)),
		"actorCount":       len(asMapSlice(actors)),
		"techniqueCount":   len(asMapSlice(techniques)),
		"toolCount":        len(asMapSlice(tools)),
		"infraCount":       len(asMapSlice(infra)),
		"iocCount":         len(asMapSlice(iocs)),
		"activeActorDist":  nonNilSlice(countActiveActorDist(knowledges, 10)),
		"matrixDist":       nonNilSlice(countBy(mapNames(knowledges, func(k kAgg) string { return k.Matrix }))),
		"malwareClassDist": nonNilSlice(countMalwareClassDist(knowledges)),
		"tacticDist":       nonNilSlice(countByFromTechniques(techniques)),
	}

	ops := buildOps(db, intelRows, knowledges)
	dashRows := buildDashRows(knowledges)
	cveRows := asMapSlice(cves)
	inWild := 0
	for _, c := range cveRows {
		if c["inWild"] == true {
			inWild++
		}
	}
	dashboard := map[string]any{
		"snapshot":                 snap,
		"operations":               ops,
		"knowledgeRows":            dashRows,
		"ransomwareKnowledgeCount": countRansomware(knowledges),
		"icsKnowledgeCount":        countICS(knowledges),
		"inWildCveCount":           inWild,
	}

	nav := []map[string]any{
		{"key": "overview", "label": "Dashboard", "icon": "📊", "count": len(knowledges)},
		{"key": "cve", "label": "漏洞利用", "icon": "🐛", "count": len(asMapSlice(cves))},
		{"key": "malware", "label": "恶意软件", "icon": "🦠", "count": len(asMapSlice(malware))},
		{"key": "actors", "label": "攻击组织", "icon": "🕵️", "count": len(asMapSlice(actors))},
		{"key": "mitre", "label": "MITRE 分布", "icon": "🎯", "count": len(asMapSlice(techniques))},
		{"key": "tools", "label": "攻击工具", "icon": "🔧", "count": len(asMapSlice(tools))},
		{"key": "infra", "label": "基础设施", "icon": "🏗️", "count": len(asMapSlice(infra))},
	}

	return &Bundle{
		Dashboard: dashboard, Nav: nav, Cve: cves, Malware: malware, Actors: actors,
		ActorProfiles: profiles, Techniques: techniques, Tools: tools, Infra: infra, Ioc: iocs,
	}, nil
}

func ref(kid int64, title string) LinkedRef {
	return LinkedRef{KnowledgeID: kid, KnowledgeTitle: title}
}

func entitySlug(s string) string {
	return url.PathEscape(s)
}

func buildCves(knowledges []kAgg) any {
	type draft struct {
		cveId, desc string
		inWild      bool
		knowledge   []LinkedRef
	}
	m := map[string]*draft{}
	for _, k := range knowledges {
		for _, v := range k.Payload.Vulnerabilities {
			cve, _ := v["cveId"].(string)
			if cve == "" {
				continue
			}
			d, ok := m[cve]
			if !ok {
				inWild, _ := v["inWild"].(bool)
				desc, _ := v["description"].(string)
				d = &draft{cveId: cve, desc: desc, inWild: inWild, knowledge: []LinkedRef{}}
				m[cve] = d
			}
			d.knowledge = append(d.knowledge, ref(k.ID, k.Title))
		}
	}
	var out []map[string]any
	for _, d := range m {
		row := finalizeRow(map[string]any{
			"cveId": d.cveId, "slug": entitySlug(d.cveId), "description": d.desc, "inWild": d.inWild,
			"knowledge": d.knowledge,
		}, d.knowledge, knowledges)
		out = append(out, applyRecency(row, d.knowledge, knowledges))
	}
	sortByRecency(out)
	return nonNilSlice(out)
}

func buildMalware(knowledges []kAgg) any {
	type draft struct {
		name, class string
		knowledge   []LinkedRef
	}
	m := map[string]*draft{}
	for _, k := range knowledges {
		for _, mal := range k.Payload.Malware {
			name, _ := mal["name"].(string)
			class := malwareclass.ClassFromMap(mal)
			key := name + "|" + class
			d, ok := m[key]
			if !ok {
				d = &draft{name: name, class: class, knowledge: []LinkedRef{}}
				m[key] = d
			}
			d.knowledge = append(d.knowledge, ref(k.ID, k.Title))
		}
	}
	var out []map[string]any
	for _, d := range m {
		slug := entitySlug(d.name + "|" + d.class)
		out = append(out, finalizeRow(map[string]any{
			"name": d.name, "class": d.class, "slug": slug, "knowledge": d.knowledge,
		}, d.knowledge, knowledges))
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["name"].(string) < out[j]["name"].(string) })
	return nonNilSlice(out)
}

func buildActors(knowledges []kAgg) (any, any) {
	type draft struct {
		name      string
		knowledge []LinkedRef
		inds      map[string]struct{}
	}
	m := map[string]*draft{}
	for _, k := range knowledges {
		for _, name := range k.Payload.ThreatActors {
			d, ok := m[name]
			if !ok {
				d = &draft{name: name, knowledge: []LinkedRef{}, inds: map[string]struct{}{}}
				m[name] = d
			}
			d.knowledge = append(d.knowledge, ref(k.ID, k.Title))
			for _, ind := range k.Industries {
				d.inds[ind] = struct{}{}
			}
		}
	}
	var list []map[string]any
	profiles := map[string]any{}
	for _, d := range m {
		slug := entitySlug(d.name)
		inds := keys(d.inds)
		row := map[string]any{
			"name": d.name, "slug": slug, "targetIndustries": inds,
			"knowledgeCount": len(d.knowledge), "knowledge": d.knowledge,
			"aliases": []string{}, "summary": nil,
		}
		list = append(list, row)
		profiles[slug] = buildActorProfile(d.name, knowledges)
	}
	sort.Slice(list, func(i, j int) bool { return list[i]["name"].(string) < list[j]["name"].(string) })
	if list == nil {
		list = []map[string]any{}
	}
	return list, profiles
}

func buildActorProfile(name string, knowledges []kAgg) map[string]any {
	var linked []kAgg
	for _, k := range knowledges {
		for _, a := range k.Payload.ThreatActors {
			if a == name {
				linked = append(linked, k)
				break
			}
		}
	}
	if len(linked) == 0 {
		return nil
	}
	base := map[string]any{
		"name": name, "slug": entitySlug(name), "aliases": []string{}, "summary": nil,
		"knowledge": []LinkedRef{}, "targetIndustries": []string{},
	}
	var refs []LinkedRef
	inds := map[string]struct{}{}
	matrices := map[string]struct{}{}
	var tools, infra, malware, vulns, techniques []map[string]any
	intelN := 0
	var first, last string
	for _, k := range linked {
		refs = append(refs, ref(k.ID, k.Title))
		for _, i := range k.Industries {
			inds[i] = struct{}{}
		}
		matrices[k.Matrix] = struct{}{}
		tools = append(tools, k.Payload.Tools...)
		infra = append(infra, k.Payload.Infrastructure...)
		malware = append(malware, k.Payload.Malware...)
		vulns = append(vulns, k.Payload.Vulnerabilities...)
		techniques = appendTechniques(techniques, k)
		intelN += k.IntelCount
		if first == "" || k.UpdatedAt < first {
			first = k.UpdatedAt
		}
		if last == "" || k.UpdatedAt > last {
			last = k.UpdatedAt
		}
	}
	base["knowledge"] = refs
	base["knowledgeCount"] = len(refs)
	base["targetIndustries"] = keys(inds)
	base["tools"] = nonNilSlice(tools)
	base["infrastructure"] = nonNilSlice(infra)
	base["malware"] = nonNilSlice(malware)
	base["vulnerabilities"] = nonNilSlice(vulns)
	base["techniques"] = nonNilSlice(techniques)
	base["intelReportCount"] = intelN
	base["matrices"] = keys(matrices)
	base["firstSeen"] = first
	base["lastSeen"] = last
	return base
}

func appendTechniques(dst []map[string]any, k kAgg) []map[string]any {
	seen := map[string]struct{}{}
	for _, t := range dst {
		if id, ok := t["techniqueId"].(string); ok {
			seen[id] = struct{}{}
		}
	}
	for _, t := range k.Payload.Techniques {
		id, _ := t["techniqueId"].(string)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		dst = append(dst, t)
	}
	return dst
}

func buildTechniques(knowledges []kAgg) any {
	m := map[string]map[string]any{}
	for _, k := range knowledges {
		for _, t := range k.Payload.Techniques {
			id, _ := t["techniqueId"].(string)
			row, ok := m[id]
			if !ok {
				tac := mitre.ResolveTechniqueTactic(id, stringField(t["tactic"]))
				row = map[string]any{
					"techniqueId": id,
					"name":        t["name"],
					"tactic":      tac,
					"matrix":      k.Matrix,
					"knowledge":   []LinkedRef{},
				}
				m[id] = row
			}
			row["knowledge"] = append(row["knowledge"].([]LinkedRef), ref(k.ID, k.Title))
		}
	}
	var out []map[string]any
	for _, row := range m {
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["techniqueId"].(string) < out[j]["techniqueId"].(string) })
	return nonNilSlice(out)
}

func buildTools(knowledges []kAgg) any {
	type draft struct {
		name, category string
		knowledge      []LinkedRef
	}
	m := map[string]*draft{}
	for _, k := range knowledges {
		for _, t := range k.Payload.Tools {
			name, _ := t["name"].(string)
			d, ok := m[name]
			if !ok {
				cat, _ := t["category"].(string)
				d = &draft{name: name, category: cat, knowledge: []LinkedRef{}}
				m[name] = d
			}
			d.knowledge = append(d.knowledge, ref(k.ID, k.Title))
		}
	}
	var out []map[string]any
	for _, d := range m {
		out = append(out, finalizeRow(map[string]any{
			"name": d.name, "category": d.category, "slug": entitySlug(d.name), "knowledge": d.knowledge,
		}, d.knowledge, knowledges))
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["name"].(string) < out[j]["name"].(string) })
	return nonNilSlice(out)
}

func buildInfra(knowledges []kAgg) any {
	type draft struct {
		typ, val, role string
		knowledge      []LinkedRef
	}
	m := map[string]*draft{}
	for _, k := range knowledges {
		for _, i := range k.Payload.Infrastructure {
			typ, _ := i["type"].(string)
			val, _ := i["value"].(string)
			key := typ + "|" + val
			d, ok := m[key]
			if !ok {
				role, _ := i["role"].(string)
				d = &draft{typ: typ, val: val, role: role, knowledge: []LinkedRef{}}
				m[key] = d
			}
			d.knowledge = append(d.knowledge, ref(k.ID, k.Title))
		}
	}
	var out []map[string]any
	for _, d := range m {
		slug := entitySlug(d.typ + "|" + d.val)
		out = append(out, finalizeRow(map[string]any{
			"type": d.typ, "value": d.val, "role": d.role, "slug": slug, "knowledge": d.knowledge,
		}, d.knowledge, knowledges))
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["value"].(string) < out[j]["value"].(string) })
	return nonNilSlice(out)
}

func buildIocs(db *store.DB, intelRows []store.IntelRow, knowledges []kAgg) any {
	m := map[string]map[string]any{}
	for _, ir := range intelRows {
		iocs, _ := db.IOCsForIntel(ir.ID)
		kids, _ := db.KnowledgeIDsForIntel(ir.ID)
		var refs []LinkedRef
		for _, kid := range kids {
			title := fmt.Sprintf("#%d", kid)
			for _, k := range knowledges {
				if k.ID == kid {
					title = k.Title
					break
				}
			}
			refs = append(refs, ref(kid, title))
		}
		for _, ioc := range iocs {
			key := ioc.Type + "|" + ioc.Value
			m[key] = map[string]any{
				"type": ioc.Type, "value": ioc.Value, "source": "intel", "knowledge": refs,
			}
		}
	}
	for _, k := range knowledges {
		for _, i := range k.Payload.Infrastructure {
			typ, _ := i["type"].(string)
			val, _ := i["value"].(string)
			key := typ + "|" + val
			if _, ok := m[key]; ok {
				continue
			}
			m[key] = map[string]any{
				"type": typ, "value": val, "source": "infrastructure",
				"knowledge": []LinkedRef{ref(k.ID, k.Title)},
			}
		}
	}
	var out []map[string]any
	for _, row := range m {
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["type"].(string) < out[j]["type"].(string) })
	return nonNilSlice(out)
}

func latestLinked(knowledge []LinkedRef, all []kAgg, field func(kAgg) string) string {
	byID := make(map[int64]string, len(all))
	for _, k := range all {
		byID[k.ID] = field(k)
	}
	latest := ""
	for _, r := range knowledge {
		if v := byID[r.KnowledgeID]; v > latest {
			latest = v
		}
	}
	return latest
}

func applyRecency(row map[string]any, knowledge []LinkedRef, all []kAgg) map[string]any {
	row["disclosedAt"] = latestLinked(knowledge, all, func(k kAgg) string { return k.DisclosedAt })
	row["updatedAt"] = latestLinked(knowledge, all, func(k kAgg) string { return k.UpdatedAt })
	return row
}

func recencyKey(row map[string]any) string {
	if d, _ := row["disclosedAt"].(string); d != "" {
		return d
	}
	u, _ := row["updatedAt"].(string)
	return u
}

func rowLabel(row map[string]any) string {
	for _, key := range []string{"cveId", "name", "techniqueId", "value"} {
		if v, _ := row[key].(string); v != "" {
			return v
		}
	}
	return ""
}

// sortByRecency orders rows newest-disclosed first; rows without disclosedAt fall back to
// updatedAt, and ties are broken by label for deterministic output.
func sortByRecency(out []map[string]any) {
	sort.Slice(out, func(i, j int) bool {
		ki, kj := recencyKey(out[i]), recencyKey(out[j])
		if ki != kj {
			return ki > kj
		}
		return rowLabel(out[i]) < rowLabel(out[j])
	})
}

func finalizeRow(row map[string]any, knowledge []LinkedRef, all []kAgg) map[string]any {
	inds := map[string]struct{}{}
	for _, r := range knowledge {
		for _, k := range all {
			if k.ID == r.KnowledgeID {
				for _, i := range k.Industries {
					inds[i] = struct{}{}
				}
			}
		}
	}
	row["targetIndustries"] = keys(inds)
	row["knowledgeCount"] = len(knowledge)
	return row
}

func buildOps(db *store.DB, intel []store.IntelRow, k []kAgg) map[string]any {
	sources := map[string]struct{}{}
	var succeeded, processing, failed, linked int
	for _, i := range intel {
		sources[i.SourceName] = struct{}{}
		switch i.IngestStatus {
		case "succeeded":
			succeeded++
		case "failed":
			failed++
		case "queued", "parsing", "extracting":
			processing++
		}
		kids, _ := db.KnowledgeIDsForIntel(i.ID)
		if len(kids) > 0 {
			linked++
		}
	}
	labels := keys(sources)
	sort.Strings(labels)
	return map[string]any{
		"knowledgeCount": len(k), "intelSourceCount": len(sources), "intelTotalCount": len(intel),
		"intelSucceededCount": succeeded, "intelProcessingCount": processing,
		"intelFailedCount": failed, "intelLinkedCount": linked, "sourceLabels": labels,
	}
}

func buildDashRows(knowledges []kAgg) []map[string]any {
	var out []map[string]any
	for _, k := range knowledges {
		hasR := false
		for _, m := range k.Payload.Malware {
			if c, _ := m["class"].(string); c == "ransomware" {
				hasR = true
			}
		}
		out = append(out, map[string]any{
			"id": k.ID, "title": k.Title, "attackMatrix": k.Matrix, "qualityScore": k.Quality,
			"intelCount": k.IntelCount, "hasRansomware": hasR, "updatedAt": k.UpdatedAt,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["updatedAt"].(string) > out[j]["updatedAt"].(string) })
	const dashboardRecentKnowledgeLimit = 10
	if len(out) > dashboardRecentKnowledgeLimit {
		out = out[:dashboardRecentKnowledgeLimit]
	}
	return out
}

func countRansomware(k []kAgg) int {
	n := 0
	for _, row := range k {
		for _, m := range row.Payload.Malware {
			if c, _ := m["class"].(string); c == "ransomware" {
				n++
				break
			}
		}
	}
	return n
}

func countICS(k []kAgg) int {
	n := 0
	for _, row := range k {
		if row.Matrix == "ICS" {
			n++
		}
	}
	return n
}

func countActiveActorDist(k []kAgg, limit int) []map[string]any {
	counts := map[string]int{}
	display := map[string]string{}
	for _, kn := range k {
		seen := map[string]struct{}{}
		for _, a := range kn.Payload.ThreatActors {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			key := strings.ToLower(a)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			if _, ok := display[key]; !ok {
				display[key] = a
			}
			counts[key]++
		}
	}
	type row struct {
		name  string
		count int
	}
	var rows []row
	for key, n := range counts {
		rows = append(rows, row{name: display[key], count: n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count != rows[j].count {
			return rows[i].count > rows[j].count
		}
		return rows[i].name < rows[j].name
	})
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"name": r.name, "count": r.count})
	}
	return out
}

func flatIndustries(k []kAgg) []string {
	var out []string
	for _, row := range k {
		out = append(out, row.Industries...)
	}
	return out
}

func mapNames(k []kAgg, fn func(kAgg) string) []string {
	var out []string
	for _, row := range k {
		out = append(out, fn(row))
	}
	return out
}

func stringField(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func countBy(names []string) []map[string]any {
	m := map[string]int{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || n == "—" || n == "-" {
			n = "未分类"
		}
		m[n]++
	}
	var out []map[string]any
	for name, count := range m {
		out = append(out, map[string]any{"name": name, "count": count})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["count"].(int) > out[j]["count"].(int) })
	return nonNilSlice(out)
}

func countMalwareClassDist(knowledges []kAgg) []map[string]any {
	var names []string
	for _, k := range knowledges {
		for _, mal := range k.Payload.Malware {
			names = append(names, malwareclass.Label(malwareclass.ClassFromMap(mal)))
		}
	}
	return countBy(names)
}

func countByFromTechniques(rows any) []map[string]any {
	list := asMapSlice(rows)
	var names []string
	for _, r := range list {
		id, _ := r["techniqueId"].(string)
		stored, _ := r["tactic"].(string)
		names = append(names, mitre.ResolveTechniqueTactic(id, stored))
	}
	return countBy(names)
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
