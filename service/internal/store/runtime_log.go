package store

import (
	"database/sql"
	"fmt"
	"strings"
)

type RuntimeLogRow struct {
	ID        int64
	CreatedAt string
	Level     string
	Source    string
	Message   string
	Detail    string
}

type RuntimeLogListParams struct {
	Level  string
	Source string
	Query  string
	Limit  int
}

func (d *DB) ensureRuntimeLogTable() error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS runtime_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at TEXT NOT NULL,
			level TEXT NOT NULL,
			source TEXT NOT NULL,
			message TEXT NOT NULL,
			detail TEXT
		)`)
	if err != nil {
		return err
	}
	_, err = d.Exec(`CREATE INDEX IF NOT EXISTS idx_runtime_log_created ON runtime_log(created_at DESC)`)
	return err
}

func (d *DB) InsertRuntimeLog(level, source, message, detail string) error {
	level = strings.TrimSpace(level)
	if level == "" {
		level = "info"
	}
	_, err := d.Exec(`
		INSERT INTO runtime_log (created_at, level, source, message, detail)
		VALUES (?, ?, ?, ?, ?)`,
		nowRFC(), level, strings.TrimSpace(source), strings.TrimSpace(message), strings.TrimSpace(detail))
	return err
}

func (d *DB) ListRuntimeLogs(p RuntimeLogListParams) ([]RuntimeLogRow, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	var conds []string
	var args []any
	if p.Level != "" {
		conds = append(conds, "level = ?")
		args = append(args, p.Level)
	}
	if p.Source != "" {
		conds = append(conds, "source = ?")
		args = append(args, p.Source)
	}
	if q := strings.TrimSpace(p.Query); q != "" {
		conds = append(conds, "(message LIKE ? OR detail LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	args = append(args, limit)
	q := fmt.Sprintf(`SELECT id, created_at, level, source, message, COALESCE(detail,'')
		FROM runtime_log %s ORDER BY id DESC LIMIT ?`, where)
	rows, err := d.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RuntimeLogRow
	for rows.Next() {
		var r RuntimeLogRow
		if err := rows.Scan(&r.ID, &r.CreatedAt, &r.Level, &r.Source, &r.Message, &r.Detail); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (d *DB) ClearRuntimeLogs() error {
	_, err := d.Exec(`DELETE FROM runtime_log`)
	return err
}

func (d *DB) RuntimeLogSources() ([]string, error) {
	rows, err := d.Query(`SELECT DISTINCT source FROM runtime_log ORDER BY source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RuntimeLogCount returns total rows (for UI header).
func (d *DB) RuntimeLogCount() (int, error) {
	var n int
	err := d.QueryRow(`SELECT COUNT(*) FROM runtime_log`).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}
