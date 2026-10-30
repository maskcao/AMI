package store

import "database/sql"

func (d *DB) ClearPendingPipelineJobs(intelID int64) error {
	_, err := d.Exec(`DELETE FROM pipeline_jobs WHERE intel_id=? AND status IN ('pending', 'failed')`, intelID)
	return err
}

// ReclaimStaleJobs re-queues jobs left "running" by a previous process exit.
func (d *DB) ReclaimStaleJobs() error {
	_, err := d.Exec(`UPDATE pipeline_jobs SET status='pending', locked_at=NULL WHERE status='running'`)
	return err
}

func (d *DB) EnqueueJob(intelID int64, stage string) error {
	_, err := d.Exec(`
		INSERT INTO pipeline_jobs (intel_id, stage, status, run_at) VALUES (?, ?, 'pending', ?)`,
		intelID, stage, nowRFC())
	return err
}

func (d *DB) ClaimNextJob() (jobID, intelID int64, stage string, err error) {
	tx, err := d.Begin()
	if err != nil {
		return 0, 0, "", err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`
		SELECT id, intel_id, stage FROM pipeline_jobs
		WHERE status='pending' AND run_at <= ?
		ORDER BY CASE stage WHEN 'extract' THEN 0 ELSE 1 END, run_at ASC, id ASC
		LIMIT 1`, nowRFC())
	if err := row.Scan(&jobID, &intelID, &stage); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, "", nil
		}
		return 0, 0, "", err
	}
	_, err = tx.Exec(`UPDATE pipeline_jobs SET status='running', locked_at=? WHERE id=?`, nowRFC(), jobID)
	if err != nil {
		return 0, 0, "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, "", err
	}
	return jobID, intelID, stage, nil
}

func (d *DB) CompleteJob(jobID int64) error {
	_, err := d.Exec(`UPDATE pipeline_jobs SET status='done', locked_at=NULL WHERE id=?`, jobID)
	return err
}

func (d *DB) FailJob(jobID int64, intelID int64, errMsg string) error {
	_, err := d.Exec(`UPDATE pipeline_jobs SET status='failed', last_error=?, attempts=attempts+1 WHERE id=?`, errMsg, jobID)
	if err != nil {
		return err
	}
	return d.SetIngestStatus(intelID, "failed", "pending", nil, errMsg, "", nil)
}
