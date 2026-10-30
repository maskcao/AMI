package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	idsFlag := flag.String("ids", "", "comma-separated intel ids")
	ticks := flag.Int("ticks", 0, "local worker ticks (0 = enqueue only, rely on running API worker)")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var ids []int64
	for _, p := range strings.Split(*idsFlag, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Fatalf("bad id %q: %v", p, err)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		log.Fatal("no ids")
	}
	if err := db.RequeueIntelStages(ids); err != nil {
		log.Fatal(err)
	}
	log.Printf("requeued %d intel ids for parse", len(ids))
	if *ticks <= 0 {
		return
	}
	w := &pipeline.Worker{DB: db, Files: artifacts.New(*dataDir)}
	for i := 0; i < *ticks; i++ {
		w.Tick()
		time.Sleep(15 * time.Millisecond)
	}
}
