package store

import (
	"database/sql"
)

type CollectSourceRow struct {
	ID            string
	Name          string
	Type          string
	Enabled       bool
	ScheduleCron  sql.NullString
	ScheduleLabel sql.NullString
	SourceLabel   string
	LastRunAt     sql.NullString
	LastStatus    sql.NullString
	LastError     sql.NullString
}

func (d *DB) ListCollectSources() ([]CollectSourceRow, error) {
	rows, err := d.Query(`SELECT id, name, type, enabled, schedule_cron, schedule_label, source_label, last_run_at, last_status, last_error FROM collect_source ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CollectSourceRow
	for rows.Next() {
		var r CollectSourceRow
		var en int
		if err := rows.Scan(&r.ID, &r.Name, &r.Type, &en, &r.ScheduleCron, &r.ScheduleLabel, &r.SourceLabel, &r.LastRunAt, &r.LastStatus, &r.LastError); err != nil {
			return nil, err
		}
		r.Enabled = en == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

type ModelProviderRow struct {
	ID         string
	Name       string
	Kind       string
	ModelID    sql.NullString
	BaseURL    sql.NullString
	APIKey     sql.NullString
	Enabled    bool
	DefaultFor sql.NullString
}

func (d *DB) scanModelProvider(sc interface {
	Scan(dest ...any) error
}) (*ModelProviderRow, error) {
	var r ModelProviderRow
	var en int
	if err := sc.Scan(&r.ID, &r.Name, &r.Kind, &r.ModelID, &r.BaseURL, &r.APIKey, &en, &r.DefaultFor); err != nil {
		return nil, err
	}
	r.Enabled = en == 1
	return &r, nil
}

func (d *DB) ListModelProviders() ([]ModelProviderRow, error) {
	rows, err := d.Query(`SELECT id, name, kind, model_id, base_url, api_key, enabled, default_for FROM model_provider ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ModelProviderRow
	for rows.Next() {
		var r ModelProviderRow
		var en int
		if err := rows.Scan(&r.ID, &r.Name, &r.Kind, &r.ModelID, &r.BaseURL, &r.APIKey, &en, &r.DefaultFor); err != nil {
			return nil, err
		}
		r.Enabled = en == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

type ParserProfileRow struct {
	ID             string
	Name           string
	Description    sql.NullString
	UsedBySources  int
}

func (d *DB) ListParserProfiles() ([]ParserProfileRow, error) {
	rows, err := d.Query(`SELECT id, name, description, used_by_sources FROM parser_profile ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ParserProfileRow
	for rows.Next() {
		var r ParserProfileRow
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.UsedBySources); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DB) SeedSettingsIfEmpty() error {
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM collect_source`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	ts := nowRFC()
	sources := []CollectSourceRow{
		{ID: "cs-1", Name: "360 Netlab 博客", Type: "paginated_list", Enabled: true, SourceLabel: "360"},
		{ID: "cs-2", Name: "CERT 通告 RSS", Type: "rss", Enabled: true, SourceLabel: "CERT"},
		{ID: "cs-3", Name: "示例公众号历史页", Type: "wechat_feed", Enabled: false, SourceLabel: "WeChat"},
	}
	for _, s := range sources {
		cron := "0 */6 * * *"
		label := "每 6 小时"
		if s.ID == "cs-2" {
			cron = "0 8 * * *"
			label = "每天 08:00"
		}
		if s.ID == "cs-3" {
			cron = "0 9 * * *"
			label = "每天 09:00"
		}
		en := 0
		if s.Enabled {
			en = 1
		}
		cfg := defaultCollectConfig(s.ID)
		_, err := d.Exec(`INSERT INTO collect_source (id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, last_run_at, last_status, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'success', ?)`,
			s.ID, s.Name, s.Type, en, cron, label, s.SourceLabel, cfg, ts, ts)
		if err != nil {
			return err
		}
	}
	_, err := d.Exec(`INSERT INTO model_provider (id, name, kind, model_id, base_url, enabled, default_for, updated_at)
		VALUES ('mp-1', '内部 LLM 网关', 'llm', 'qwen2.5-72b-instruct', 'https://llm.internal.example/v1', 1, 'parse, enrich', ?)`, ts)
	if err != nil {
		return err
	}
	profiles := []struct{ id, name, desc string; used int }{
		{"pp-1", "generic_html_v1", "通用安全博客 HTML 正文", 2},
		{"pp-2", "wechat_article_html_v1", "微信公众号文章排版", 1},
	}
	for _, p := range profiles {
		_, err = d.Exec(`INSERT INTO parser_profile (id, name, description, used_by_sources, config_json, updated_at) VALUES (?, ?, ?, ?, '{}', ?)`,
			p.id, p.name, p.desc, p.used, ts)
		if err != nil {
			return err
		}
	}
	return nil
}

func defaultCollectConfig(id string) string {
	switch id {
	case "cs-1":
		return `{"list_url":"https://blog.netlab.360.com/tag/botnet/","pagination":{"type":"path","path_template":"https://blog.netlab.360.com/tag/botnet/page/{page}/","start_page":2,"max_pages":3,"stop_on_empty_list":true},"link_extract":{"must_contain":"blog.netlab.360.com","exclude_contains":["/tag/","/page/","/feed"],"prefer_path_regex":"/20[0-9]{2}/"},"article_base":"https://blog.netlab.360.com","max_items_per_run":10,"delay_ms":500}`
	case "cs-2":
		return `{"feed_url":"https://www.cisa.gov/cybersecurity-advisories/all.xml","max_items_per_run":5}`
	default:
		return `{}`
	}
}

// EnsureCollectSourceConfigs backfills config_json when still empty (existing DBs).
func (d *DB) EnsureCollectSourceConfigs() error {
	rows, err := d.Query(`SELECT id, config_json FROM collect_source`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, cfg string
		if err := rows.Scan(&id, &cfg); err != nil {
			return err
		}
		if cfg != "" && cfg != "{}" {
			continue
		}
		def := defaultCollectConfig(id)
		if def == "{}" {
			continue
		}
		if _, err := d.Exec(`UPDATE collect_source SET config_json=?, updated_at=? WHERE id=?`, def, nowRFC(), id); err != nil {
			return err
		}
	}
	return rows.Err()
}
