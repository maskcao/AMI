package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	dryRun := flag.Bool("dry-run", false, "report only")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, title, summary, payload_json FROM attack_knowledge`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var updated, skipped int
	for rows.Next() {
		var id int64
		var title, summary, payloadJSON string
		if err := rows.Scan(&id, &title, &summary, &payloadJSON); err != nil {
			log.Fatalf("scan: %v", err)
		}
		pl, err := store.ParsePayload(payloadJSON)
		if err != nil {
			skipped++
			continue
		}
		actors := pipeline.MergeThreatActorsExported(pl.ThreatActors, title, summary)
		if len(actors) == len(pl.ThreatActors) {
			same := true
			for i := range actors {
				if actors[i] != pl.ThreatActors[i] {
					same = false
					break
				}
			}
			if same {
				skipped++
				continue
			}
		}
		if len(actors) == 0 {
			skipped++
			continue
		}
		if *dryRun {
			log.Printf("knowledge %d: %v", id, actors)
			updated++
			continue
		}
		pl.ThreatActors = actors
		b, _ := json.Marshal(pl)
		if _, err := db.Exec(`UPDATE attack_knowledge SET payload_json=? WHERE id=?`, string(b), id); err != nil {
			log.Printf("knowledge %d: %v", id, err)
			continue
		}
		updated++
	}
	log.Printf("done: updated=%d skipped=%d dry_run=%v", updated, skipped, *dryRun)
}
