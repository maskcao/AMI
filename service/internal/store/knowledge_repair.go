package store

import (
	"fmt"
	"strings"
)

type linkedIntel struct {
	ID      int64
	Title   string
	Summary string
}

// SplitKnowledgeByIntelTitle splits one knowledge row linked to multiple distinct intel titles.
// The earliest intel (lowest id) keeps the original knowledge id with its title; others get new knowledge rows.
func (d *DB) SplitKnowledgeByIntelTitle(knowledgeID int64) (created int, err error) {
	rows, err := d.Query(`
		SELECT ir.id, ir.title, COALESCE(ir.summary_snippet, '')
		FROM intel_report ir
		INNER JOIN knowledge_intel_link ikl ON ikl.intel_id = ir.id
		WHERE ikl.knowledge_id = ?
		ORDER BY ir.id ASC`, knowledgeID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var items []linkedIntel
	for rows.Next() {
		var it linkedIntel
		if err := rows.Scan(&it.ID, &it.Title, &it.Summary); err != nil {
			return created, err
		}
		it.Title = strings.TrimSpace(it.Title)
		it.Summary = cleanSummarySnippet(it.Summary)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return created, err
	}
	if len(items) <= 1 {
		return 0, nil
	}

	titles := map[string]struct{}{}
	for _, it := range items {
		titles[it.Title] = struct{}{}
	}
	if len(titles) <= 1 {
		return 0, nil
	}

	keep := items[0]
	existing, err := d.GetKnowledge(knowledgeID)
	if err != nil {
		return 0, err
	}
	pl := KnowledgePayload{Techniques: []map[string]any{}}
	if existing != nil {
		pl, _ = ParsePayload(existing.PayloadJSON)
	}
	if err := d.UpdateKnowledgeSummary(knowledgeID, keep.Title, keep.Summary, pl); err != nil {
		return 0, err
	}
	if err := d.ExecSetIntelKnowledgeAction(keep.ID, "updated"); err != nil {
		return 0, err
	}

	for _, it := range items[1:] {
		newPl := KnowledgePayload{Techniques: []map[string]any{}}
		kid, err := d.CreateKnowledge(it.Title, it.Summary, "Enterprise", []string{"互联网"}, newPl)
		if err != nil {
			return created, fmt.Errorf("intel %d: create knowledge: %w", it.ID, err)
		}
		if _, err := d.Exec(`DELETE FROM knowledge_intel_link WHERE knowledge_id=? AND intel_id=?`, knowledgeID, it.ID); err != nil {
			return created, err
		}
		if err := d.LinkIntel(kid, it.ID, "primary"); err != nil {
			return created, err
		}
		if err := d.ExecSetIntelKnowledgeAction(it.ID, "created"); err != nil {
			return created, err
		}
		created++
	}
	if err := d.refreshKnowledgeIntelCount(knowledgeID); err != nil {
		return created, err
	}
	return created, nil
}

func (d *DB) refreshKnowledgeIntelCount(knowledgeID int64) error {
	_, err := d.Exec(`
		UPDATE attack_knowledge SET intel_count = (SELECT COUNT(*) FROM knowledge_intel_link WHERE knowledge_id=?), updated_at=? WHERE id=?`,
		knowledgeID, nowRFC(), knowledgeID)
	return err
}

func cleanSummarySnippet(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	var kept []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "奇安信威胁情报中心" || strings.HasPrefix(line, "返回 TI") || line == "RESEARCH" {
			continue
		}
		if strings.Contains(line, "数 据 驱 动 安 全") {
			continue
		}
		kept = append(kept, line)
	}
	if len(kept) == 0 {
		return strings.TrimSpace(s)
	}
	out := strings.Join(kept, "\n")
	if len(out) > 512 {
		out = out[:512] + "…"
	}
	return out
}

func (d *DB) ExecSetIntelKnowledgeAction(intelID int64, action string) error {
	_, err := d.Exec(`UPDATE intel_report SET knowledge_action=?, updated_at=? WHERE id=?`, action, nowRFC(), intelID)
	return err
}

// ListOversizedKnowledgeByDistinctTitles returns knowledge ids with multiple intel and distinct titles.
func (d *DB) ListOversizedKnowledgeByDistinctTitles() ([]int64, error) {
	rows, err := d.Query(`
		SELECT ak.id
		FROM attack_knowledge ak
		INNER JOIN knowledge_intel_link ikl ON ikl.knowledge_id = ak.id
		INNER JOIN intel_report ir ON ir.id = ikl.intel_id
		GROUP BY ak.id
		HAVING COUNT(DISTINCT ikl.intel_id) > 1 AND COUNT(DISTINCT ir.title) > 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
