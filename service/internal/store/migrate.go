package store

import (
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed sql/*.sql
var migrationFS embed.FS

func (d *DB) Migrate() error {
	entries, err := migrationFS.ReadDir("sql")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		b, err := migrationFS.ReadFile("sql/" + name)
		if err != nil {
			return err
		}
		if _, err := d.Exec(string(b)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}
