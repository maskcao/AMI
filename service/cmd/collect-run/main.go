// collect-run runs a collect source once without pilot/batch gates (ops/dev).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/ingest"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	sourceID := flag.String("id", "", "collect source id (required)")
	flag.Parse()
	if *sourceID == "" {
		fmt.Fprintln(os.Stderr, "usage: collect-run -data ./data -id <collect-source-id>")
		os.Exit(2)
	}
	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(*dataDir)
	res, err := ingest.RunCollectSource(db, files, *sourceID)
	if err != nil {
		log.Printf("collect error: %v", err)
	}
	if res != nil {
		fmt.Printf("imported=%d skipped=%d failed=%d\n", res.Imported, res.Skipped, res.Failed)
	}
	if err != nil {
		os.Exit(1)
	}
}
