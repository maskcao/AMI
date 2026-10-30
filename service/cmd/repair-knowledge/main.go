package main

import (
	"flag"
	"log"

	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	knowledgeID := flag.Int64("knowledge-id", 0, "repair only this knowledge id (0 = scan all)")
	dryRun := flag.Bool("dry-run", false, "list only, do not split")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var ids []int64
	if *knowledgeID > 0 {
		ids = []int64{*knowledgeID}
	} else {
		ids, err = db.ListOversizedKnowledgeByDistinctTitles()
		if err != nil {
			log.Fatalf("list: %v", err)
		}
	}
	if len(ids) == 0 {
		log.Println("no oversized knowledge entries to repair")
		return
	}

	totalCreated := 0
	for _, id := range ids {
		row, _ := db.GetKnowledge(id)
		title := ""
		if row != nil {
			title = row.Title
		}
		if *dryRun {
			log.Printf("would split knowledge %d (%q)", id, title)
			continue
		}
		n, err := db.SplitKnowledgeByIntelTitle(id)
		if err != nil {
			log.Printf("knowledge %d: %v", id, err)
			continue
		}
		log.Printf("knowledge %d (%q): split -> %d new knowledge rows", id, title, n)
		totalCreated += n
	}
	log.Printf("done: %d knowledge entries processed, %d new rows created", len(ids), totalCreated)
}
