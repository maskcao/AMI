package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory (contains ami.db)")
	dryRun := flag.Bool("dry-run", false, "print changes without updating DB")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	files := artifacts.New(*dataDir)
	rows, err := db.Query(`SELECT id, title, source_url, collect_source_name FROM intel_report
		WHERE title = '奇安信威胁情报中心' OR source_url LIKE '%ti.qianxin.com%' OR source_url LIKE '%ti.360.net/blog%'`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var updated, skipped, failed int
	for rows.Next() {
		var id int64
		var title, sourceURL, collectName string
		if err := rows.Scan(&id, &title, &sourceURL, &collectName); err != nil {
			log.Fatalf("scan: %v", err)
		}
		rawPath := filepath.Join(*dataDir, "store", "intel", fmt.Sprintf("%d", id), "raw", "001.html")
		raw, err := os.ReadFile(rawPath)
		if err != nil {
			raw, err = files.ReadRaw(id, 1)
		}
		if err != nil {
			log.Printf("intel %d: no raw html: %v", id, err)
			failed++
			continue
		}
		opts := pipeline.TitleOptsForIntel(db, collectName, sourceURL)
		newTitle := pipeline.ExtractTitleFromHTML(string(raw), opts)
		if newTitle == "" {
			newTitle = pipeline.ResolveIntelTitle(title, sourceURL, raw, opts)
		}
		if newTitle == "" || newTitle == title {
			skipped++
			continue
		}
		if pipeline.NormalizeTitle(newTitle) == pipeline.NormalizeTitle(title) {
			skipped++
			continue
		}
		log.Printf("intel %d: %q -> %q", id, title, newTitle)
		if *dryRun {
			updated++
			continue
		}
		if err := db.UpdateIntelTitle(id, newTitle); err != nil {
			log.Printf("intel %d: update: %v", id, err)
			failed++
			continue
		}
		updated++
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}
	log.Printf("done: updated=%d skipped=%d failed=%d dry_run=%v", updated, skipped, failed, *dryRun)

	dup, _ := countDuplicateTitles(db)
	if dup > 0 {
		log.Printf("warn: %d title values still shared by multiple intel rows (review knowledge merge)", dup)
	}
}

func countDuplicateTitles(db *store.DB) (int, error) {
	row := db.QueryRow(`SELECT COUNT(*) FROM (
		SELECT title FROM intel_report GROUP BY title HAVING COUNT(*) > 1
	)`)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
