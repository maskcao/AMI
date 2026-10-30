package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

const qianxinURLPattern = "%ti.qianxin.com%"

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	urlPattern := flag.String("url-pattern", qianxinURLPattern, "SQL LIKE pattern for intel.source_url")
	knowledgeID := flag.Int64("knowledge-id", 0, "only delete+reprocess intel linked to this knowledge id")
	runPipeline := flag.Bool("run-pipeline", true, "drain parse+extract jobs after enqueue")
	maxTicks := flag.Int("max-ticks", 800, "max worker ticks when -run-pipeline")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	files := artifacts.New(*dataDir)

	var queued int
	if *knowledgeID > 0 {
		n, err := db.DeleteKnowledgeByIDAndReextract(*knowledgeID)
		if err != nil {
			log.Fatalf("delete knowledge %d: %v", *knowledgeID, err)
		}
		queued = n
		log.Printf("deleted knowledge %d, requeued parse for %d intel", *knowledgeID, n)
	} else {
		kDel, intelN, err := db.DeleteKnowledgeLinkedToIntelURL(*urlPattern)
		if err != nil {
			log.Fatalf("delete knowledge: %v", err)
		}
		log.Printf("deleted %d attack_knowledge rows, reset %d intel reports", kDel, intelN)
		queued, err = db.EnqueueIntelReprocess(*urlPattern)
		if err != nil {
			log.Fatalf("enqueue: %v", err)
		}
		log.Printf("enqueued parse for %d intel rows", queued)
	}

	if !*runPipeline || queued == 0 {
		return
	}

	worker := &pipeline.Worker{DB: db, Files: files}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	done := 0
	for tick := 0; tick < *maxTicks; tick++ {
		select {
		case <-ctx.Done():
			log.Fatalf("pipeline timeout after %d ticks", tick)
		default:
		}
		before := countPending(db)
		worker.Tick()
		after := countPending(db)
		if before > 0 && after == 0 {
			done++
			if done >= 3 {
				break
			}
		} else {
			done = 0
		}
		time.Sleep(20 * time.Millisecond)
	}

	var succeeded, pending, other int
	row := db.QueryRow(`SELECT
		SUM(CASE WHEN ingest_status='succeeded' THEN 1 ELSE 0 END),
		SUM(CASE WHEN ingest_status IN ('queued','parsing','extracting') THEN 1 ELSE 0 END),
		COUNT(*)
		FROM intel_report WHERE source_url LIKE ?`, *urlPattern)
	_ = row.Scan(&succeeded, &pending, &other)
	log.Printf("pipeline drain finished: succeeded=%d in-flight=%d total=%d", succeeded, pending, other)

	kn, _ := countKnowledge(db, *urlPattern)
	log.Printf("attack_knowledge linked to pattern: %d", kn)
}

func countPending(db *store.DB) int {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM pipeline_jobs WHERE status='pending'`).Scan(&n)
	return n
}

func countKnowledge(db *store.DB, pattern string) (int, error) {
	var n int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT ak.id) FROM attack_knowledge ak
		INNER JOIN knowledge_intel_link ikl ON ikl.knowledge_id = ak.id
		INNER JOIN intel_report ir ON ir.id = ikl.intel_id
		WHERE ir.source_url LIKE ?`, pattern).Scan(&n)
	return n, err
}
