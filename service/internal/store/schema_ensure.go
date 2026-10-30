package store

import (
	"database/sql"
	"strings"
)

func (d *DB) ensureSchemaPatches() error {
	if err := d.ensureModelProviderAPIKey(); err != nil {
		return err
	}
	if err := d.ensureCollectSourcePilotColumns(); err != nil {
		return err
	}
	if err := d.ensureRuntimeLogTable(); err != nil {
		return err
	}
	if err := d.ensureIncidentReportTime(); err != nil {
		return err
	}
	if err := d.ensureIntelStopRequested(); err != nil {
		return err
	}
	return d.ensureKnowledgeDupColumn()
}

func (d *DB) ensureIntelStopRequested() error {
	var name string
	err := d.QueryRow(`SELECT name FROM pragma_table_info('intel_report') WHERE name = 'stop_requested'`).Scan(&name)
	if err == sql.ErrNoRows {
		_, err = d.Exec(`ALTER TABLE intel_report ADD COLUMN stop_requested INTEGER NOT NULL DEFAULT 0`)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (d *DB) ensureKnowledgeDupColumn() error {
	var name string
	err := d.QueryRow(`SELECT name FROM pragma_table_info('attack_knowledge') WHERE name = 'dup_candidates_json'`).Scan(&name)
	if err == sql.ErrNoRows {
		_, err = d.Exec(`ALTER TABLE attack_knowledge ADD COLUMN dup_candidates_json TEXT NOT NULL DEFAULT ''`)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (d *DB) ensureIncidentReportTime() error {
	var name string
	err := d.QueryRow(`SELECT name FROM pragma_table_info('incident_report') WHERE name = 'report_time'`).Scan(&name)
	if err == sql.ErrNoRows {
		_, err = d.Exec(`ALTER TABLE incident_report ADD COLUMN report_time TEXT NOT NULL DEFAULT ''`)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}

func (d *DB) ensureModelProviderAPIKey() error {
	var name string
	err := d.QueryRow(`SELECT name FROM pragma_table_info('model_provider') WHERE name = 'api_key'`).Scan(&name)
	if err == sql.ErrNoRows {
		_, err = d.Exec(`ALTER TABLE model_provider ADD COLUMN api_key TEXT`)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return err
		}
	}
	return nil
}
