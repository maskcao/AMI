package store

import (
	"database/sql"
	"time"
)

type IncidentReport struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Filename    string `json:"filename"`
	Ext         string `json:"ext"`
	Mime        string `json:"mime"`
	SizeBytes   int64  `json:"sizeBytes"`
	SHA256      string `json:"sha256"`
	StoragePath string `json:"-"`
	Note        string `json:"note"`
	ReportTime  string `json:"reportTime"`
	UploadedAt  string `json:"uploadedAt"`
}

func (d *DB) CreateIncidentReport(title, filename, ext, mime string, size int64, sha, storagePath, reportTime string) (*IncidentReport, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := d.Exec(
		`INSERT INTO incident_report (title, filename, ext, mime, size_bytes, sha256, storage_path, report_time, uploaded_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		title, filename, ext, mime, size, sha, storagePath, reportTime, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return d.GetIncidentReport(id)
}

func (d *DB) GetIncidentReport(id int64) (*IncidentReport, error) {
	row := d.QueryRow(
		`SELECT id, title, filename, ext, mime, size_bytes, sha256, storage_path, note, report_time, uploaded_at
		 FROM incident_report WHERE id = ?`, id,
	)
	return scanIncident(row)
}

func (d *DB) ListIncidentReports() ([]*IncidentReport, error) {
	rows, err := d.Query(
		`SELECT id, title, filename, ext, mime, size_bytes, sha256, storage_path, note, report_time, uploaded_at
		 FROM incident_report ORDER BY uploaded_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*IncidentReport
	for rows.Next() {
		it, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (d *DB) UpdateIncidentReportTime(id int64, reportTime string) error {
	res, err := d.Exec(`UPDATE incident_report SET report_time = ? WHERE id = ?`, reportTime, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (d *DB) DeleteIncidentReport(id int64) error {
	res, err := d.Exec(`DELETE FROM incident_report WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanIncident(row interface{ Scan(...any) error }) (*IncidentReport, error) {
	var it IncidentReport
	if err := row.Scan(&it.ID, &it.Title, &it.Filename, &it.Ext, &it.Mime, &it.SizeBytes, &it.SHA256, &it.StoragePath, &it.Note, &it.ReportTime, &it.UploadedAt); err != nil {
		return nil, err
	}
	return &it, nil
}
