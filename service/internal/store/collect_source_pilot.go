package store

import (
	"database/sql"
	"strings"
)

const collectSourceSelectSQL = `id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json,
	last_run_at, last_status, last_error, batch_unlocked, pilot_article_passed_at, pilot_list_passed_at`

func scanCollectSourceFull(sc interface {
	Scan(dest ...any) error
}, r *CollectSourceFull) error {
	var en, unlocked int
	var articleAt, listAt sql.NullString
	if err := sc.Scan(
		&r.ID, &r.Name, &r.Type, &en, &r.ScheduleCron, &r.ScheduleLabel, &r.SourceLabel, &r.ConfigJSON,
		&r.LastRunAt, &r.LastStatus, &r.LastError, &unlocked, &articleAt, &listAt,
	); err != nil {
		return err
	}
	r.Enabled = en == 1
	r.BatchUnlocked = unlocked == 1
	r.PilotArticlePassedAt = articleAt
	r.PilotListPassedAt = listAt
	return nil
}

func (d *DB) ensureCollectSourcePilotColumns() error {
	cols := []struct{ name, ddl string }{
		{"batch_unlocked", `ALTER TABLE collect_source ADD COLUMN batch_unlocked INTEGER NOT NULL DEFAULT 1`},
		{"pilot_article_passed_at", `ALTER TABLE collect_source ADD COLUMN pilot_article_passed_at TEXT`},
		{"pilot_list_passed_at", `ALTER TABLE collect_source ADD COLUMN pilot_list_passed_at TEXT`},
	}
	for _, c := range cols {
		var name string
		err := d.QueryRow(`SELECT name FROM pragma_table_info('collect_source') WHERE name = ?`, c.name).Scan(&name)
		if err == sql.ErrNoRows {
			if _, err := d.Exec(c.ddl); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) SetCollectSourcePilotArticlePassed(id string) error {
	ts := nowRFC()
	_, err := d.Exec(`UPDATE collect_source SET pilot_article_passed_at=?, batch_unlocked=0, updated_at=? WHERE id=?`, ts, ts, id)
	return err
}

func (d *DB) SetCollectSourcePilotListPassed(id string) error {
	ts := nowRFC()
	_, err := d.Exec(`UPDATE collect_source SET pilot_list_passed_at=?, updated_at=? WHERE id=?`, ts, ts, id)
	return err
}

func (d *DB) SetCollectSourceBatchUnlocked(id string, unlocked bool) error {
	v := 0
	if unlocked {
		v = 1
	}
	_, err := d.Exec(`UPDATE collect_source SET batch_unlocked=?, updated_at=? WHERE id=?`, v, nowRFC(), id)
	return err
}

func (d *DB) ResetCollectSourcePilot(id string) error {
	_, err := d.Exec(`UPDATE collect_source SET batch_unlocked=0, pilot_article_passed_at=NULL, pilot_list_passed_at=NULL, updated_at=? WHERE id=?`,
		nowRFC(), id)
	return err
}

func (d *DB) GetIntelIDBySourceURL(sourceURL string) (int64, error) {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return 0, nil
	}
	var id int64
	err := d.QueryRow(`SELECT id FROM intel_report WHERE source_url=? AND archived_at IS NULL ORDER BY id DESC LIMIT 1`, sourceURL).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (d *DB) PilotReadyForBatch(src *CollectSourceFull) (bool, string) {
	if src == nil {
		return false, "source not found"
	}
	if !src.PilotArticlePassedAt.Valid || src.PilotArticlePassedAt.String == "" {
		return false, "未完成单条试采或未通过质检"
	}
	if ingestTypeNeedsListPilot(src.Type) {
		if !src.PilotListPassedAt.Valid || src.PilotListPassedAt.String == "" {
			return false, "未完成分页/列表试采"
		}
	}
	if !src.BatchUnlocked {
		return false, "未解锁批量运行"
	}
	return true, ""
}

func ingestTypeNeedsListPilot(typ string) bool {
	switch typ {
	case "paginated_list", "threat_portal", "scheduled_url", "wechat_feed":
		return true
	default:
		return false
	}
}

func (d *DB) MarkListPilotSkippedForDirectTypes(id string, typ string) error {
	if ingestTypeNeedsListPilot(typ) {
		return nil
	}
	ts := nowRFC()
	_, err := d.Exec(`UPDATE collect_source SET pilot_list_passed_at=?, updated_at=? WHERE id=? AND (pilot_list_passed_at IS NULL OR pilot_list_passed_at='')`,
		ts, ts, id)
	return err
}

// CanUnlockBatch checks article (+ list if required) pilot timestamps before unlock.
func (d *DB) CanUnlockBatch(src *CollectSourceFull) (bool, string) {
	if src == nil {
		return false, "source not found"
	}
	if !src.PilotArticlePassedAt.Valid || src.PilotArticlePassedAt.String == "" {
		return false, "请先完成并通过单条试采"
	}
	if ingestTypeNeedsListPilot(src.Type) {
		if !src.PilotListPassedAt.Valid || src.PilotListPassedAt.String == "" {
			return false, "请先完成并通过分页/列表试采"
		}
	}
	return true, ""
}

func (d *DB) UpdateCollectSourceWithPilotReset(id string, p CollectSourcePatch, resetPilot bool) error {
	if resetPilot {
		if err := d.ResetCollectSourcePilot(id); err != nil {
			return err
		}
	}
	return d.UpdateCollectSource(id, p)
}
