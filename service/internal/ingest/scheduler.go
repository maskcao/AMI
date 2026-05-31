package ingest

import (
	"context"
	"log"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/store"
)

type Scheduler struct {
	DB      *store.DB
	Files   *artifacts.Store
	MinGap  time.Duration
}

func (s *Scheduler) Run(ctx context.Context, tick time.Duration) {
	if tick <= 0 {
		tick = 5 * time.Minute
	}
	if s.MinGap <= 0 {
		s.MinGap = time.Hour
	}
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tickDue()
		}
	}
}

func (s *Scheduler) tickDue() {
	sources, err := s.DB.ListEnabledCollectSources()
	if err != nil {
		log.Printf("ingest list sources: %v", err)
		return
	}
	now := time.Now()
	for _, src := range sources {
		if src.Type == "manual_upload" {
			continue
		}
		lastRun := ""
		if src.LastRunAt.Valid {
			lastRun = src.LastRunAt.String
		}
		if !s.isDue(lastRun, now) {
			continue
		}
		id := src.ID
		go func() {
			opslog.Info("scheduler", "scheduled collect tick", "sourceId="+id)
			if _, err := RunCollectSource(s.DB, s.Files, id); err != nil {
				log.Printf("ingest scheduled %s: %v", id, err)
			}
		}()
	}
}

func (s *Scheduler) isDue(lastRun string, now time.Time) bool {
	if lastRun == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, lastRun)
	if err != nil {
		return true
	}
	return now.Sub(t) >= s.MinGap
}
