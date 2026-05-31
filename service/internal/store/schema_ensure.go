package store

import (
	"database/sql"
	"strings"
)

func (d *DB) ensureSchemaPatches() error {
	if err := d.ensureModelProviderAPIKey(); err != nil {
		return err
	}
	return d.ensureRuntimeLogTable()
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
