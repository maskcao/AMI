package store

import "database/sql"

const (
	IntelStopStopped  = "stopped"
	IntelStopStopping = "stopping"
	IntelStopSkipped  = "skipped"
)

// RequestIntelStop cancels a queued item immediately, or asks a running stage to
// stop at its next checkpoint (the current parse/extract step still finishes).
func (d *DB) RequestIntelStop(id int64) (string, error) {
	var status string
	err := d.QueryRow(`SELECT ingest_status FROM intel_report WHERE id=?`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return IntelStopSkipped, nil
	}
	if err != nil {
		return "", err
	}
	switch status {
	case "queued":
		if _, err := d.Exec(`DELETE FROM pipeline_jobs WHERE intel_id=? AND status IN ('pending','failed')`, id); err != nil {
			return "", err
		}
		if _, err := d.Exec(`UPDATE intel_report SET ingest_status='stopped', stop_requested=1, pipeline_error=NULL, updated_at=? WHERE id=?`, nowRFC(), id); err != nil {
			return "", err
		}
		return IntelStopStopped, nil
	case "parsing", "extracting":
		if _, err := d.Exec(`UPDATE intel_report SET stop_requested=1, updated_at=? WHERE id=?`, nowRFC(), id); err != nil {
			return "", err
		}
		return IntelStopStopping, nil
	default:
		return IntelStopSkipped, nil
	}
}

// IntelStopRequested is polled by the pipeline worker before and after each stage.
func (d *DB) IntelStopRequested(id int64) (bool, error) {
	var flag int
	var status string
	err := d.QueryRow(`SELECT stop_requested, ingest_status FROM intel_report WHERE id=?`, id).Scan(&flag, &status)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return flag == 1 || status == "stopped", nil
}

// ApplyIntelStop clears pending follow-up jobs and marks the item stopped unless
// the stage already reached a terminal status.
func (d *DB) ApplyIntelStop(id int64) error {
	if _, err := d.Exec(`DELETE FROM pipeline_jobs WHERE intel_id=? AND status IN ('pending','failed')`, id); err != nil {
		return err
	}
	if _, err := d.Exec(`UPDATE intel_report SET stop_requested=0, updated_at=? WHERE id=?`, nowRFC(), id); err != nil {
		return err
	}
	_, err := d.Exec(`UPDATE intel_report SET ingest_status='stopped', pipeline_error=NULL, updated_at=? WHERE id=? AND ingest_status IN ('queued','parsing','extracting')`, nowRFC(), id)
	return err
}

func (d *DB) ClearIntelStop(id int64) error {
	_, err := d.Exec(`UPDATE intel_report SET stop_requested=0 WHERE id=?`, id)
	return err
}

// IntelIDsForStop returns ids for batch stop: "active" covers queued + running.
func (d *DB) IntelIDsForStop(scope string) ([]int64, error) {
	var rows *sql.Rows
	var err error
	if scope == "active" {
		rows, err = d.Query(`SELECT id FROM intel_report WHERE ingest_status IN ('queued','parsing','extracting') AND archived_at IS NULL ORDER BY id`)
	} else {
		rows, err = d.Query(`SELECT id FROM intel_report WHERE ingest_status='queued' AND archived_at IS NULL ORDER BY id`)
	}
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
