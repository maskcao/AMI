package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ami/ami/service/internal/api"
	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/config"
	"github.com/ami/ami/service/internal/ingest"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/seed"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	seedFlag := flag.Bool("seed", false, "seed demo data if database is empty")
	flag.Parse()

	cfg := config.Load()
	if *seedFlag {
		cfg.Seed = true
	}

	db, err := store.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	opslog.Init(db)

	files := artifacts.New(cfg.DataDir)
	if err := db.SeedSettingsIfEmpty(); err != nil {
		log.Fatalf("settings seed: %v", err)
	}
	if err := db.EnsureCollectSourceConfigs(); err != nil {
		log.Fatalf("collect config: %v", err)
	}
	if cfg.Seed {
		if err := seed.Run(db, files); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	worker := &pipeline.Worker{DB: db, Files: files}
	if err := db.ReclaimStaleJobs(); err != nil {
		log.Printf("reclaim stale jobs: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx, time.Duration(config.WorkerIntervalSec())*time.Second)

	ingestSch := &ingest.Scheduler{
		DB: db, Files: files,
		MinGap: time.Duration(config.IngestMinGapSec()) * time.Second,
	}
	go ingestSch.Run(ctx, time.Duration(config.IngestTickSec())*time.Second)

	if cfg.Seed {
		drainPipeline(worker, 60)
	}

	srv := api.New(db, files)
	httpSrv := &http.Server{Addr: cfg.Addr, Handler: srv}

	opslog.Info("system", "AMI API started", cfg.Addr)

	go func() {
		log.Printf("AMI API listening on %s (data: %s)", cfg.Addr, cfg.DataDir)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func drainPipeline(w *pipeline.Worker, max int) {
	for i := 0; i < max; i++ {
		w.Tick()
		time.Sleep(50 * time.Millisecond)
	}
}
