package store

import "database/sql"

type LinkedIntelBrief struct {
	ID                int64          `json:"id"`
	Title             string         `json:"title"`
	SourceName        string         `json:"sourceName"`
	SourceURL         string         `json:"sourceUrl"`
	CollectSourceName string         `json:"collectSourceName"`
	DisclosedAt       sql.NullString `json:"-"`
	Role              string         `json:"role"`
	HasRaw            bool           `json:"hasRaw"`
	HasParsed         bool           `json:"hasParsed"`
}

func (d *DB) LinkedIntelForKnowledge(knowledgeID int64) ([]LinkedIntelBrief, error) {
	rows, err := d.Query(`
		SELECT ir.id, ir.title, ir.source_name, ir.source_url, ir.collect_source_name,
			ir.disclosed_at, kil.role, ir.has_raw, ir.has_parsed
		FROM knowledge_intel_link kil
		JOIN intel_report ir ON ir.id = kil.intel_id
		WHERE kil.knowledge_id = ?
		ORDER BY CASE kil.role WHEN 'primary' THEN 0 ELSE 1 END, ir.updated_at DESC`,
		knowledgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LinkedIntelBrief
	for rows.Next() {
		var b LinkedIntelBrief
		var hasRaw, hasParsed int
		if err := rows.Scan(&b.ID, &b.Title, &b.SourceName, &b.SourceURL, &b.CollectSourceName,
			&b.DisclosedAt, &b.Role, &hasRaw, &hasParsed); err != nil {
			return nil, err
		}
		b.HasRaw = hasRaw == 1
		b.HasParsed = hasParsed == 1
		out = append(out, b)
	}
	return out, rows.Err()
}

func (d *DB) LatestPipelineIntel(knowledgeID int64) (*IntelRow, error) {
	row := d.QueryRow(`
		SELECT ir.id FROM knowledge_intel_link kil
		JOIN intel_report ir ON ir.id = kil.intel_id
		WHERE kil.knowledge_id = ? AND ir.ingest_status = 'succeeded'
		ORDER BY ir.updated_at DESC LIMIT 1`, knowledgeID)
	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return d.GetIntel(id)
}

func (d *DB) FiguresForKnowledge(knowledgeID int64) ([]IntelFigure, error) {
	rows, err := d.Query(`
		SELECT f.id, f.intel_id, f.title, f.caption, f.kind, f.src, f.technique_ids_json
		FROM intel_figure f
		WHERE f.intel_id IN (SELECT intel_id FROM knowledge_intel_link WHERE knowledge_id = ?)`,
		knowledgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFigures(rows)
}
