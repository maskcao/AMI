package main

import (
	"encoding/json"
	"flag"
	"log"

	"github.com/ami/ami/service/internal/mitre"
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

	rows, err := db.Query(`SELECT id, payload_json FROM attack_knowledge`)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var updated, skipped int
	for rows.Next() {
		var id int64
		var payloadJSON string
		if err := rows.Scan(&id, &payloadJSON); err != nil {
			log.Fatalf("scan: %v", err)
		}
		pl, err := store.ParsePayload(payloadJSON)
		if err != nil || len(pl.Techniques) == 0 {
			skipped++
			continue
		}
		changed := false
		for i, t := range pl.Techniques {
			tid, _ := t["techniqueId"].(string)
			stored, _ := t["tactic"].(string)
			resolved := mitre.ResolveTechniqueTactic(tid, stored)
			if resolved == stored {
				continue
			}
			if resolved == "" {
				continue
			}
			t["tactic"] = resolved
			pl.Techniques[i] = t
			changed = true
		}
		if !changed {
			skipped++
			continue
		}
		if *dryRun {
			log.Printf("knowledge %d: would update %d techniques", id, len(pl.Techniques))
			updated++
			continue
		}
		b, _ := json.Marshal(pl)
		if _, err := db.Exec(`UPDATE attack_knowledge SET payload_json=? WHERE id=?`, string(b), id); err != nil {
			log.Printf("knowledge %d: %v", id, err)
			continue
		}
		updated++
	}
	log.Printf("done: updated=%d skipped=%d dry_run=%v", updated, skipped, *dryRun)
}
