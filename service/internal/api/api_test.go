package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/ami/ami/service/internal/api"
	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/seed"
	"github.com/ami/ami/service/internal/store"
)

func TestAPIIntelAndKnowledge(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(dir)
	if err := seed.Run(db, files); err != nil {
		t.Fatal(err)
	}
	w := pipeline.Worker{DB: db, Files: files}
	for i := 0; i < 60; i++ {
		w.Tick()
		time.Sleep(20 * time.Millisecond)
	}

	srv := api.New(db, files)

	// list intel — DB only
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/intel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list intel: %d %s", rec.Code, rec.Body.String())
	}

	// detail without full original in default response
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/intel/4203", nil))
	var intel map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &intel)
	if intel["ingestStatus"] != "succeeded" {
		t.Fatalf("intel status: %v", intel["ingestStatus"])
	}

	// original on demand
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/intel/4203/original", nil))
	if rec.Code != http.StatusOK || len(rec.Body.Bytes()) < 10 {
		t.Fatalf("original: %d len=%d", rec.Code, rec.Body.Len())
	}

	// knowledge detail with linked intel
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/1001", nil))
	var kn map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &kn)
	linked, _ := kn["linkedIntel"].([]any)
	if len(linked) < 1 {
		t.Fatalf("expected linkedIntel, got %v", kn["linkedIntel"])
	}

	// patch + publish
	body, _ := json.Marshal(map[string]any{"summary": "updated summary for test"})
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/knowledge/1001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/1001/publish", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("publish: %d %s", rec.Code, rec.Body.String())
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "store", "corpus", "knowledge", "1001", "snapshot-*.jsonl"))
	if len(matches) == 0 {
		t.Fatal("corpus snapshot not written on publish")
	}
}
