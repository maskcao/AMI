package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CollectSourceFull struct {
	CollectSourceRow
	ConfigJSON string
}

func (d *DB) GetCollectSource(id string) (*CollectSourceFull, error) {
	row := d.QueryRow(`SELECT id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, last_run_at, last_status, last_error
		FROM collect_source WHERE id = ?`, id)
	var r CollectSourceFull
	var en int
	if err := row.Scan(&r.ID, &r.Name, &r.Type, &en, &r.ScheduleCron, &r.ScheduleLabel, &r.SourceLabel, &r.ConfigJSON,
		&r.LastRunAt, &r.LastStatus, &r.LastError); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	r.Enabled = en == 1
	return &r, nil
}

type CollectSourcePatch struct {
	Name          *string
	Enabled       *bool
	ScheduleCron  *string
	ScheduleLabel *string
	SourceLabel   *string
	ConfigJSON    *string
}

func (d *DB) UpdateCollectSource(id string, p CollectSourcePatch) error {
	cur, err := d.GetCollectSource(id)
	if err != nil || cur == nil {
		return fmt.Errorf("collect source not found: %s", id)
	}
	name := cur.Name
	en := 0
	if cur.Enabled {
		en = 1
	}
	cron := cur.ScheduleCron
	label := cur.ScheduleLabel
	srcLabel := cur.SourceLabel
	cfg := cur.ConfigJSON

	if p.Name != nil {
		name = strings.TrimSpace(*p.Name)
	}
	if p.Enabled != nil {
		if *p.Enabled {
			en = 1
		} else {
			en = 0
		}
	}
	if p.ScheduleCron != nil {
		cron = sql.NullString{String: strings.TrimSpace(*p.ScheduleCron), Valid: *p.ScheduleCron != ""}
	}
	if p.ScheduleLabel != nil {
		label = sql.NullString{String: strings.TrimSpace(*p.ScheduleLabel), Valid: *p.ScheduleLabel != ""}
	}
	if p.SourceLabel != nil {
		srcLabel = strings.TrimSpace(*p.SourceLabel)
	}
	if p.ConfigJSON != nil {
		if !json.Valid([]byte(*p.ConfigJSON)) {
			return fmt.Errorf("invalid config json")
		}
		cfg = *p.ConfigJSON
	}
	_, err = d.Exec(`UPDATE collect_source SET name=?, enabled=?, schedule_cron=?, schedule_label=?, source_label=?, config_json=?, updated_at=? WHERE id=?`,
		name, en, cron, label, srcLabel, cfg, nowRFC(), id)
	return err
}

type CollectSourceCreate struct {
	Name          string
	Type          string
	Enabled       bool
	ScheduleCron  string
	ScheduleLabel string
	SourceLabel   string
	ConfigJSON    string
}

func (d *DB) CreateCollectSource(c CollectSourceCreate) (string, error) {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return "", fmt.Errorf("name required")
	}
	typ := strings.TrimSpace(c.Type)
	if typ == "" {
		return "", fmt.Errorf("type required")
	}
	cfg := strings.TrimSpace(c.ConfigJSON)
	if cfg == "" {
		cfg = "{}"
	}
	if !json.Valid([]byte(cfg)) {
		return "", fmt.Errorf("invalid config json")
	}
	id := fmt.Sprintf("cs-%d", time.Now().UnixNano())
	en := 0
	if c.Enabled {
		en = 1
	}
	cron := sql.NullString{String: strings.TrimSpace(c.ScheduleCron), Valid: strings.TrimSpace(c.ScheduleCron) != ""}
	label := sql.NullString{String: strings.TrimSpace(c.ScheduleLabel), Valid: strings.TrimSpace(c.ScheduleLabel) != ""}
	srcLabel := strings.TrimSpace(c.SourceLabel)
	_, err := d.Exec(`INSERT INTO collect_source (id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, last_status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'never', ?)`,
		id, name, typ, en, cron, label, srcLabel, cfg, nowRFC())
	return id, err
}

func (d *DB) DeleteCollectSource(id string) error {
	res, err := d.Exec(`DELETE FROM collect_source WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("collect source not found: %s", id)
	}
	return nil
}

func (d *DB) SetCollectSourceRunResult(id, status, errMsg string) error {
	ts := nowRFC()
	var errVal sql.NullString
	if errMsg != "" {
		errVal = sql.NullString{String: errMsg, Valid: true}
	}
	_, err := d.Exec(`UPDATE collect_source SET last_run_at=?, last_status=?, last_error=?, updated_at=? WHERE id=?`,
		ts, status, errVal, ts, id)
	return err
}

func (d *DB) ListEnabledCollectSources() ([]CollectSourceFull, error) {
	rows, err := d.Query(`SELECT id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, last_run_at, last_status, last_error
		FROM collect_source WHERE enabled = 1 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CollectSourceFull
	for rows.Next() {
		var r CollectSourceFull
		var en int
		if err := rows.Scan(&r.ID, &r.Name, &r.Type, &en, &r.ScheduleCron, &r.ScheduleLabel, &r.SourceLabel, &r.ConfigJSON,
			&r.LastRunAt, &r.LastStatus, &r.LastError); err != nil {
			return nil, err
		}
		r.Enabled = en == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

type ModelProviderPatch struct {
	Name       *string
	ModelID    *string
	BaseURL    *string
	APIKey     *string // non-nil: set; empty string clears
	Enabled    *bool
	DefaultFor *string
}

type ModelProviderCreate struct {
	Name       string
	Kind       string
	ModelID    string
	BaseURL    string
	APIKey     string
	DefaultFor string
	Enabled    bool
}

func (d *DB) GetModelProvider(id string) (*ModelProviderRow, error) {
	row := d.QueryRow(`SELECT id, name, kind, model_id, base_url, api_key, enabled, default_for FROM model_provider WHERE id = ?`, id)
	r, err := d.scanModelProvider(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (d *DB) CreateModelProvider(c ModelProviderCreate) (string, error) {
	id := fmt.Sprintf("mp-%d", time.Now().UnixNano())
	kind := c.Kind
	if kind == "" {
		kind = "llm"
	}
	en := 0
	if c.Enabled {
		en = 1
	}
	modelID := sql.NullString{String: strings.TrimSpace(c.ModelID), Valid: strings.TrimSpace(c.ModelID) != ""}
	baseURL := sql.NullString{String: strings.TrimSpace(c.BaseURL), Valid: strings.TrimSpace(c.BaseURL) != ""}
	apiKey := sql.NullString{String: strings.TrimSpace(c.APIKey), Valid: strings.TrimSpace(c.APIKey) != ""}
	defFor := sql.NullString{String: strings.TrimSpace(c.DefaultFor), Valid: strings.TrimSpace(c.DefaultFor) != ""}
	_, err := d.Exec(`INSERT INTO model_provider (id, name, kind, model_id, base_url, api_key, enabled, default_for, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, strings.TrimSpace(c.Name), kind, modelID, baseURL, apiKey, en, defFor, nowRFC())
	return id, err
}

func (d *DB) DeleteModelProvider(id string) error {
	res, err := d.Exec(`DELETE FROM model_provider WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("model provider not found: %s", id)
	}
	return nil
}

func (d *DB) UpdateModelProvider(id string, p ModelProviderPatch) error {
	cur, err := d.GetModelProvider(id)
	if err != nil || cur == nil {
		return fmt.Errorf("model provider not found: %s", id)
	}
	name := cur.Name
	modelID := cur.ModelID
	baseURL := cur.BaseURL
	defFor := cur.DefaultFor
	en := 0
	if cur.Enabled {
		en = 1
	}
	if p.Name != nil {
		name = strings.TrimSpace(*p.Name)
	}
	if p.ModelID != nil {
		modelID = sql.NullString{String: strings.TrimSpace(*p.ModelID), Valid: *p.ModelID != ""}
	}
	if p.BaseURL != nil {
		baseURL = sql.NullString{String: strings.TrimSpace(*p.BaseURL), Valid: *p.BaseURL != ""}
	}
	if p.DefaultFor != nil {
		defFor = sql.NullString{String: strings.TrimSpace(*p.DefaultFor), Valid: *p.DefaultFor != ""}
	}
	if p.Enabled != nil {
		if *p.Enabled {
			en = 1
		} else {
			en = 0
		}
	}
	apiKey := cur.APIKey
	if p.APIKey != nil {
		apiKey = sql.NullString{String: strings.TrimSpace(*p.APIKey), Valid: strings.TrimSpace(*p.APIKey) != ""}
	}
	_, err = d.Exec(`UPDATE model_provider SET name=?, model_id=?, base_url=?, api_key=?, enabled=?, default_for=?, updated_at=? WHERE id=?`,
		name, modelID, baseURL, apiKey, en, defFor, nowRFC(), id)
	return err
}

type ParserProfileCreate struct {
	Name        string
	Description string
}

func (d *DB) CreateParserProfile(c ParserProfileCreate) (string, error) {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return "", fmt.Errorf("name required")
	}
	id := fmt.Sprintf("pp-%d", time.Now().UnixNano())
	desc := sql.NullString{String: strings.TrimSpace(c.Description), Valid: strings.TrimSpace(c.Description) != ""}
	_, err := d.Exec(`INSERT INTO parser_profile (id, name, description, used_by_sources, config_json, updated_at) VALUES (?, ?, ?, 0, '{}', ?)`,
		id, name, desc, nowRFC())
	return id, err
}

func (d *DB) GetParserProfile(id string) (*ParserProfileRow, error) {
	row := d.QueryRow(`SELECT id, name, description, used_by_sources FROM parser_profile WHERE id = ?`, id)
	var r ParserProfileRow
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &r.UsedBySources); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

type ParserProfilePatch struct {
	Name        *string
	Description *string
}

func (d *DB) UpdateParserProfile(id string, p ParserProfilePatch) error {
	cur, err := d.GetParserProfile(id)
	if err != nil || cur == nil {
		return fmt.Errorf("parser profile not found: %s", id)
	}
	name := cur.Name
	desc := cur.Description
	if p.Name != nil {
		name = strings.TrimSpace(*p.Name)
	}
	if p.Description != nil {
		desc = sql.NullString{String: strings.TrimSpace(*p.Description), Valid: true}
	}
	_, err = d.Exec(`UPDATE parser_profile SET name=?, description=?, updated_at=? WHERE id=?`, name, desc, nowRFC(), id)
	return err
}

func (d *DB) DeleteParserProfile(id string) error {
	var used int
	if err := d.QueryRow(`SELECT used_by_sources FROM parser_profile WHERE id = ?`, id).Scan(&used); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parser profile not found: %s", id)
		}
		return err
	}
	if used > 0 {
		return fmt.Errorf("parser profile is referenced by %d collect source(s)", used)
	}
	res, err := d.Exec(`DELETE FROM parser_profile WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("parser profile not found: %s", id)
	}
	return nil
}

func (d *DB) DefaultLLMProvider() (*ModelProviderRow, error) {
	rows, err := d.ListModelProviders()
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if !r.Enabled || r.Kind != "llm" {
			continue
		}
		if r.DefaultFor.Valid && strings.Contains(r.DefaultFor.String, "enrich") {
			return &r, nil
		}
	}
	for _, r := range rows {
		if r.Enabled && r.Kind == "llm" {
			return &r, nil
		}
	}
	return nil, nil
}
