package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory (contains ami.db)")
	dryRun := flag.Bool("dry-run", false, "print changes without updating DB")
	ids := flag.String("ids", "", "comma-separated intel ids (empty = all with raw html)")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	files := artifacts.New(*dataDir)
	q := `SELECT id FROM intel_report WHERE has_raw=1`
	args := []any{}
	if *ids != "" {
		q += ` AND id IN (` + *ids + `)`
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var updated, skipped int
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("scan: %v", err)
		}
		raw, err := files.ReadRaw(id, 1)
		if err != nil {
			log.Printf("intel %d: read raw: %v", id, err)
			continue
		}
		disclosed := pipeline.ExtractDisclosedAtFromHTML(string(raw))
		if disclosed == "" {
			skipped++
			continue
		}
		if *dryRun {
			fmt.Printf("intel %d disclosed_at=%s\n", id, disclosed)
			updated++
			continue
		}
		if err := db.UpdateIntelDisclosedAt(id, disclosed); err != nil {
			log.Printf("intel %d: update: %v", id, err)
			continue
		}
		updated++
	}
	log.Printf("done: updated=%d skipped(no date)=%d dry_run=%v", updated, skipped, *dryRun)
}
