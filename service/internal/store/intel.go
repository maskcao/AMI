package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type IntelRow struct {
	ID                 int64
	Title              string
	SourceURL          string
	SourceName         string
	CollectSourceName  string
	DisclosedAt        sql.NullString
	CollectedAt        string
	UpdatedAt          string
	SummarySnippet     sql.NullString
	BodyPreview        sql.NullString
	IngestStatus       string
	KnowledgeAction    string
	ExtractConfidence  sql.NullFloat64
	PipelineError      sql.NullString
	ParseVersion       sql.NullString
	EnrichVersion      sql.NullString
	HasRaw             bool
	HasParsed          bool
	AIExtractJSON      sql.NullString
	StopRequested      bool
}

type IntelIOC struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence *string `json:"confidence,omitempty"`
	ExpiresAt  *string `json:"expiresAt"`
}

type IntelFigure struct {
	ID           string   `json:"id"`
	IntelID      int64    `json:"intelId"`
	Title        string   `json:"title"`
	Caption      string   `json:"caption,omitempty"`
	Kind         string   `json:"kind"`
	Src          string   `json:"src"`
	TechniqueIDs []string `json:"techniqueIds,omitempty"`
}

type AIExtract struct {
	Summary             string   `json:"summary"`
	SuggestedTechniques []string `json:"suggestedTechniques"`
	SuggestedMalware    []string `json:"suggestedMalware"`
	SuggestedCves       []string `json:"suggestedCves"`
	Tools               []string `json:"tools,omitempty"`
	Infrastructure      []string `json:"infrastructure,omitempty"`
	MalwareClasses      []string `json:"malwareClasses,omitempty"`
	QualityHints        []string `json:"qualityHints,omitempty"`
	QualityOK           bool     `json:"qualityOk,omitempty"`
}

type IntelListParams struct {
	SourceName   string
	KnowledgeID  int64
	IngestStatus string
	SortOrder    string
	Limit        int
	Offset       int
}

func nowRFC() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) IntelExistsBySourceURL(sourceURL string) (bool, error) {
	if sourceURL == "" {
		return false, nil
	}
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM intel_report WHERE source_url = ? AND archived_at IS NULL`, sourceURL).Scan(&n)
	return n > 0, err
}

func (d *DB) CreateIntel(title, sourceURL, sourceName, collectSource string) (int64, error) {
	ts := nowRFC()
	res, err := d.Exec(`
		INSERT INTO intel_report (title, source_url, source_name, collect_source_name, collected_at, updated_at, ingest_status, knowledge_action)
		VALUES (?, ?, ?, ?, ?, ?, 'queued', 'pending')`,
		title, sourceURL, sourceName, collectSource, ts, ts)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateIntelAfterRaw(id int64, bodyPreview string) error {
	_, err := d.Exec(`
		UPDATE intel_report SET has_raw=1, body_preview=?, updated_at=? WHERE id=?`,
		truncate(bodyPreview, 2048), nowRFC(), id)
	return err
}

func (d *DB) UpdateIntelTitle(id int64, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil
	}
	_, err := d.Exec(`UPDATE intel_report SET title=?, updated_at=? WHERE id=?`, truncate(title, 512), nowRFC(), id)
	return err
}

func (d *DB) UpdateIntelDisclosedAt(id int64, disclosedAt string) error {
	disclosedAt = strings.TrimSpace(disclosedAt)
	if disclosedAt == "" {
		return nil
	}
	_, err := d.Exec(`UPDATE intel_report SET disclosed_at=?, updated_at=? WHERE id=?`, disclosedAt, nowRFC(), id)
	return err
}

func (d *DB) UpdateIntelAfterParse(id int64, snippet, parseVersion, mdRel string) error {
	ts := nowRFC()
	_, err := d.Exec(`
		UPDATE intel_report SET has_parsed=1, summary_snippet=?, parse_version=?, updated_at=? WHERE id=?`,
		truncate(snippet, 512), parseVersion, ts, id)
	if err != nil {
		return err
	}
	_, err = d.Exec(`
		INSERT OR REPLACE INTO intel_parsed (intel_id, version, body_md_path, body_txt_path, parsed_at, parser_profile_id)
		VALUES (?, ?, ?, '', ?, 'html-v1')`, id, "v1", mdRel, ts)
	return err
}

// ParsedLLMInputPath returns body.md for extract; falls back to legacy body.txt.
func (d *DB) ParsedLLMInputPath(intelID int64) (string, error) {
	if p, err := d.ParsedMdPath(intelID); err == nil && p != "" {
		return p, nil
	}
	return d.ParsedTxtPath(intelID)
}

func (d *DB) SetIngestStatus(id int64, status, knowledgeAction string, conf *float64, pipelineErr string, enrichVersion string, ai *AIExtract) error {
	var aiJSON sql.NullString
	if ai != nil {
		b, _ := json.Marshal(ai)
		aiJSON = sql.NullString{String: string(b), Valid: true}
	}
	var errVal sql.NullString
	if pipelineErr != "" {
		errVal = sql.NullString{String: pipelineErr, Valid: true}
	}
	var confVal sql.NullFloat64
	if conf != nil {
		confVal = sql.NullFloat64{Float64: *conf, Valid: true}
	}
	_, err := d.Exec(`
		UPDATE intel_report SET ingest_status=?, knowledge_action=?, extract_confidence=?, pipeline_error=?, enrich_version=?, ai_extract_json=?, updated_at=? WHERE id=?`,
		status, knowledgeAction, confVal, errVal, enrichVersion, aiJSON, nowRFC(), id)
	return err
}

func (d *DB) InsertArtifact(intelID int64, seq int, role, storagePath, sha256 string, size int64, mime string) error {
	_, err := d.Exec(`
		INSERT OR REPLACE INTO intel_artifact (intel_id, seq, role, storage_path, sha256, size_bytes, mime, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		intelID, seq, role, storagePath, sha256, size, mime, nowRFC())
	return err
}

func (d *DB) InsertIOC(intelID int64, ioc IntelIOC) error {
	_, err := d.Exec(`INSERT INTO intel_ioc (intel_id, type, value, confidence) VALUES (?, ?, ?, ?)`,
		intelID, ioc.Type, ioc.Value, nullStr(ioc.Confidence))
	return err
}

func (d *DB) InsertFigure(f IntelFigure) error {
	tids, _ := json.Marshal(f.TechniqueIDs)
	_, err := d.Exec(`
		INSERT OR REPLACE INTO intel_figure (id, intel_id, title, caption, kind, src, technique_ids_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.IntelID, f.Title, f.Caption, f.Kind, f.Src, string(tids))
	return err
}

func (d *DB) ListIntel(p IntelListParams) ([]IntelRow, int, error) {
	if p.Limit <= 0 {
		p.Limit = 100
	}
	var conds []string
	var args []any
	if p.SourceName != "" {
		conds = append(conds, "source_name = ?")
		args = append(args, p.SourceName)
	}
	if p.IngestStatus != "" {
		conds = append(conds, "ingest_status = ?")
		args = append(args, p.IngestStatus)
	}
	if p.KnowledgeID > 0 {
		conds = append(conds, "id IN (SELECT intel_id FROM knowledge_intel_link WHERE knowledge_id = ?)")
		args = append(args, p.KnowledgeID)
	}
	conds = append(conds, "archived_at IS NULL")
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM intel_report `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := "DESC"
	if p.SortOrder == "asc" {
		order = "ASC"
	}
	q := fmt.Sprintf(`SELECT id, title, source_url, source_name, collect_source_name, disclosed_at, collected_at, updated_at,
		summary_snippet, body_preview, ingest_status, knowledge_action, extract_confidence, pipeline_error,
		parse_version, enrich_version, has_raw, has_parsed, ai_extract_json, stop_requested
		FROM intel_report %s ORDER BY updated_at %s LIMIT ? OFFSET ?`, where, order)
	args = append(args, p.Limit, p.Offset)
	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []IntelRow
	for rows.Next() {
		r, err := scanIntel(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (d *DB) GetIntel(id int64) (*IntelRow, error) {
	row := d.QueryRow(`SELECT id, title, source_url, source_name, collect_source_name, disclosed_at, collected_at, updated_at,
		summary_snippet, body_preview, ingest_status, knowledge_action, extract_confidence, pipeline_error,
		parse_version, enrich_version, has_raw, has_parsed, ai_extract_json, stop_requested FROM intel_report WHERE id=?`, id)
	r, err := scanIntelRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func scanIntel(rows *sql.Rows) (IntelRow, error) {
	var r IntelRow
	var hasRaw, hasParsed, stopReq int
	err := rows.Scan(&r.ID, &r.Title, &r.SourceURL, &r.SourceName, &r.CollectSourceName, &r.DisclosedAt,
		&r.CollectedAt, &r.UpdatedAt, &r.SummarySnippet, &r.BodyPreview, &r.IngestStatus, &r.KnowledgeAction,
		&r.ExtractConfidence, &r.PipelineError, &r.ParseVersion, &r.EnrichVersion, &hasRaw, &hasParsed, &r.AIExtractJSON,
		&stopReq)
	r.HasRaw = hasRaw == 1
	r.HasParsed = hasParsed == 1
	r.StopRequested = stopReq == 1
	return r, err
}

func scanIntelRow(row *sql.Row) (*IntelRow, error) {
	var r IntelRow
	var hasRaw, hasParsed, stopReq int
	err := row.Scan(&r.ID, &r.Title, &r.SourceURL, &r.SourceName, &r.CollectSourceName, &r.DisclosedAt,
		&r.CollectedAt, &r.UpdatedAt, &r.SummarySnippet, &r.BodyPreview, &r.IngestStatus, &r.KnowledgeAction,
		&r.ExtractConfidence, &r.PipelineError, &r.ParseVersion, &r.EnrichVersion, &hasRaw, &hasParsed, &r.AIExtractJSON,
		&stopReq)
	if err != nil {
		return nil, err
	}
	r.HasRaw = hasRaw == 1
	r.HasParsed = hasParsed == 1
	r.StopRequested = stopReq == 1
	return &r, nil
}

func (d *DB) KnowledgeIDsForIntel(intelID int64) ([]int64, error) {
	rows, err := d.Query(`SELECT knowledge_id FROM knowledge_intel_link WHERE intel_id=? ORDER BY role`, intelID)
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

func (d *DB) IOCsForIntel(intelID int64) ([]IntelIOC, error) {
	rows, err := d.Query(`SELECT type, value, confidence, expires_at FROM intel_ioc WHERE intel_id=?`, intelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IntelIOC
	for rows.Next() {
		var i IntelIOC
		var conf, exp sql.NullString
		if err := rows.Scan(&i.Type, &i.Value, &conf, &exp); err != nil {
			return nil, err
		}
		if conf.Valid {
			i.Confidence = &conf.String
		}
		if exp.Valid {
			i.ExpiresAt = &exp.String
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (d *DB) FiguresForIntel(intelID int64) ([]IntelFigure, error) {
	rows, err := d.Query(`SELECT id, intel_id, title, caption, kind, src, technique_ids_json FROM intel_figure WHERE intel_id=?`, intelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFigures(rows)
}

func scanFigures(rows *sql.Rows) ([]IntelFigure, error) {
	var out []IntelFigure
	for rows.Next() {
		var f IntelFigure
		var cap sql.NullString
		var tids sql.NullString
		if err := rows.Scan(&f.ID, &f.IntelID, &f.Title, &cap, &f.Kind, &f.Src, &tids); err != nil {
			return nil, err
		}
		if cap.Valid {
			f.Caption = cap.String
		}
		if tids.Valid && tids.String != "" {
			_ = json.Unmarshal([]byte(tids.String), &f.TechniqueIDs)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (d *DB) IOCCount(intelID int64) (int, error) {
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM intel_ioc WHERE intel_id=?`, intelID).Scan(&n)
	return n, err
}

func (d *DB) PriorIntelForMerge(excludeID int64) ([]struct {
	Title        string
	KnowledgeIDs []int64
}, error) {
	rows, err := d.Query(`
		SELECT ir.id, ir.title FROM intel_report ir
		WHERE ir.id != ? AND ir.ingest_status = 'succeeded'`, excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Title        string
		KnowledgeIDs []int64
	}
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		kids, err := d.KnowledgeIDsForIntel(id)
		if err != nil {
			return nil, err
		}
		out = append(out, struct {
			Title        string
			KnowledgeIDs []int64
		}{title, kids})
	}
	return out, rows.Err()
}

func (d *DB) ParsedTxtPath(intelID int64) (string, error) {
	var p sql.NullString
	err := d.QueryRow(`SELECT body_txt_path FROM intel_parsed WHERE intel_id=? AND version='v1'`, intelID).Scan(&p)
	if err != nil || !p.Valid {
		return "", err
	}
	return p.String, nil
}

func (d *DB) ParsedMdPath(intelID int64) (string, error) {
	var p sql.NullString
	err := d.QueryRow(`SELECT body_md_path FROM intel_parsed WHERE intel_id=? AND version='v1'`, intelID).Scan(&p)
	if err != nil || !p.Valid {
		return "", err
	}
	return p.String, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func nullStr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}
