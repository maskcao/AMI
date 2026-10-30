package main

import (
	"flag"
	"log"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

// Standalone pipeline drain (use when API worker appears stuck or for catch-up).
func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	ticks := flag.Int("ticks", 2000, "worker tick iterations")
	interval := flag.Duration("interval", 25*time.Millisecond, "sleep between ticks")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	w := &pipeline.Worker{DB: db, Files: artifacts.New(*dataDir)}
	for i := 0; i < *ticks; i++ {
		w.Tick()
		if i > 0 && i%100 == 0 {
			log.Printf("tick %d", i)
		}
		time.Sleep(*interval)
	}
	log.Printf("drain finished after %d ticks", *ticks)
}
