package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type KnowledgePayload struct {
	Techniques      []map[string]any `json:"techniques"`
	Tools           []map[string]any `json:"tools"`
	Infrastructure  []map[string]any `json:"infrastructure"`
	Vulnerabilities []map[string]any `json:"vulnerabilities"`
	Malware         []map[string]any `json:"malware"`
	ThreatActors    []string         `json:"threatActors"`
	Procedures      []map[string]any `json:"procedures"`
	QualityHints    []string         `json:"qualityHints"`
	Pipeline        map[string]any   `json:"pipeline"`
	LinkedIntel     []map[string]any `json:"linkedIntel"`
	EvidenceFigures []map[string]any `json:"evidenceFigures"`
}

type KnowledgeRow struct {
	ID                   int64
	Title                string
	Summary              sql.NullString
	Status               string
	NeedsReview          bool
	AttackMatrix         string
	TargetIndustriesJSON string
	QualityScore         int
	IntelCount           int
	EarliestDisclosedAt  sql.NullString
	UpdatedAt            string
	PayloadJSON          string
	EnrichVersion        sql.NullString
	DupCandidatesJSON    string
}

const knowledgeEarliestDisclosedSQL = `(
	SELECT MIN(ir.disclosed_at) FROM knowledge_intel_link kil
	INNER JOIN intel_report ir ON ir.id = kil.intel_id
	WHERE kil.knowledge_id = attack_knowledge.id
	  AND ir.disclosed_at IS NOT NULL AND TRIM(ir.disclosed_at) <> ''
)`

type KnowledgeListParams struct {
	Status       string
	NeedsReview  *bool
	Query        string
	Matrix       string
	Industry     string
	MalwareClass string
	LLMFailed    bool
	Limit        int
	Offset       int
}

func (d *DB) CreateKnowledge(title, summary, matrix string, industries []string, payload KnowledgePayload) (int64, error) {
	ind, _ := json.Marshal(industries)
	pl, _ := json.Marshal(payload)
	ts := nowRFC()
	dups, _ := json.Marshal(d.findDupCandidates(0, title, payload))
	res, err := d.Exec(`
		INSERT INTO attack_knowledge (title, summary, status, needs_review, attack_matrix, target_industries_json, quality_score, intel_count, updated_at, payload_json, dup_candidates_json)
		VALUES (?, ?, 'draft', 1, ?, ?, 0, 0, ?, ?, ?)`,
		title, summary, matrix, string(ind), ts, string(pl), string(dups))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateKnowledgeSummary(id int64, title, summary string, payload KnowledgePayload) error {
	pl, _ := json.Marshal(payload)
	dups, _ := json.Marshal(d.findDupCandidates(id, title, payload))
	_, err := d.Exec(`
		UPDATE attack_knowledge SET title=?, summary=?, payload_json=?, updated_at=?, needs_review=1, dup_candidates_json=? WHERE id=?`,
		title, summary, string(pl), nowRFC(), string(dups), id)
	return err
}

func (d *DB) LinkIntel(knowledgeID, intelID int64, role string) error {
	_, err := d.Exec(`
		INSERT OR REPLACE INTO knowledge_intel_link (knowledge_id, intel_id, role, linked_at) VALUES (?, ?, ?, ?)`,
		knowledgeID, intelID, role, nowRFC())
	if err != nil {
		return err
	}
	_, err = d.Exec(`
		UPDATE attack_knowledge SET intel_count = (SELECT COUNT(*) FROM knowledge_intel_link WHERE knowledge_id=?), updated_at=? WHERE id=?`,
		knowledgeID, nowRFC(), knowledgeID)
	return err
}

func (d *DB) ListKnowledge(p KnowledgeListParams) ([]KnowledgeRow, int, error) {
	if p.Limit <= 0 {
		p.Limit = 100
	}
	var conds []string
	var args []any
	if p.Status != "" {
		conds = append(conds, "status = ?")
		args = append(args, p.Status)
	}
	if p.NeedsReview != nil {
		v := 0
		if *p.NeedsReview {
			v = 1
		}
		conds = append(conds, "needs_review = ?")
		args = append(args, v)
	}
	if p.Query != "" {
		conds = append(conds, "(title LIKE ? OR summary LIKE ?)")
		q := "%" + p.Query + "%"
		args = append(args, q, q)
	}
	if p.Matrix != "" {
		conds = append(conds, "attack_matrix = ?")
		args = append(args, p.Matrix)
	}
	if p.Industry != "" {
		conds = append(conds, "target_industries_json LIKE ?")
		args = append(args, "%\""+p.Industry+"\"%")
	}
	if p.MalwareClass != "" {
		conds = append(conds, `payload_json LIKE ?`)
		args = append(args, "%\"class\":\""+p.MalwareClass+"\"%")
	}
	if p.LLMFailed {
		conds = append(conds, `EXISTS (SELECT 1 FROM knowledge_intel_link kil WHERE kil.knowledge_id = attack_knowledge.id)
			AND NOT EXISTS (SELECT 1 FROM knowledge_intel_link kil
				INNER JOIN intel_report ir ON ir.id = kil.intel_id
				WHERE kil.knowledge_id = attack_knowledge.id AND ir.enrich_version LIKE 'llm-v2%')`)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM attack_knowledge `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := fmt.Sprintf(`SELECT id, title, summary, status, needs_review, attack_matrix, target_industries_json, quality_score, intel_count, %s, updated_at, payload_json, dup_candidates_json,
		(SELECT ir.enrich_version FROM knowledge_intel_link kil
			INNER JOIN intel_report ir ON ir.id = kil.intel_id
			WHERE kil.knowledge_id = attack_knowledge.id
			ORDER BY ir.updated_at DESC LIMIT 1) AS enrich_version
		FROM attack_knowledge %s ORDER BY updated_at DESC LIMIT ? OFFSET ?`, knowledgeEarliestDisclosedSQL, where)
	args = append(args, p.Limit, p.Offset)
	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []KnowledgeRow
	for rows.Next() {
		r, err := scanKnowledge(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (d *DB) GetKnowledge(id int64) (*KnowledgeRow, error) {
	row := d.QueryRow(`SELECT id, title, summary, status, needs_review, attack_matrix, target_industries_json, quality_score, intel_count, updated_at, payload_json, dup_candidates_json FROM attack_knowledge WHERE id=?`, id)
	var r KnowledgeRow
	var nr int
	err := row.Scan(&r.ID, &r.Title, &r.Summary, &r.Status, &nr, &r.AttackMatrix, &r.TargetIndustriesJSON, &r.QualityScore, &r.IntelCount, &r.UpdatedAt, &r.PayloadJSON, &r.DupCandidatesJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	r.NeedsReview = nr == 1
	return &r, err
}

func scanKnowledge(rows *sql.Rows) (KnowledgeRow, error) {
	var r KnowledgeRow
	var nr int
	err := rows.Scan(&r.ID, &r.Title, &r.Summary, &r.Status, &nr, &r.AttackMatrix, &r.TargetIndustriesJSON, &r.QualityScore, &r.IntelCount, &r.EarliestDisclosedAt, &r.UpdatedAt, &r.PayloadJSON, &r.DupCandidatesJSON, &r.EnrichVersion)
	r.NeedsReview = nr == 1
	return r, err
}

func ParsePayload(jsonStr string) (KnowledgePayload, error) {
	var p KnowledgePayload
	if jsonStr == "" {
		return p, nil
	}
	err := json.Unmarshal([]byte(jsonStr), &p)
	return p, err
}
