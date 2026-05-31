package store

import (
	"os"
	"path/filepath"
)

// PurgeAllData removes all OLTP rows and artifact files under data/store.
func (d *DB) PurgeAllData() error {
	tables := []string{
		"pipeline_jobs",
		"knowledge_intel_link",
		"intel_ioc",
		"intel_artifact",
		"intel_report",
		"attack_knowledge",
		"collect_source",
		"model_provider",
		"parser_profile",
		"runtime_log",
	}
	for _, t := range tables {
		if _, err := d.Exec("DELETE FROM " + t); err != nil {
			return err
		}
	}
	root := filepath.Join(d.dataDir, "store")
	_ = os.RemoveAll(root)
	return os.MkdirAll(root, 0o755)
}
