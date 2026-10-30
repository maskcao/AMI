package ingest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/store"
)

func TestImportListQueue_skipsKnownThenImportsNew(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a1/", "/a2/", "/a3/":
			fmt.Fprintf(w, `<html><head><title>%s</title></head><body>ok</body></html>`, r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(dir)
	ts := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO collect_source (id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, updated_at)
		VALUES ('cs-q', 'Q', 'paginated_list', 1, '0 * * * *', 'h', 'T', '{}', ?)`, ts)
	if err != nil {
		t.Fatal(err)
	}

	queue := []listLink{
		{URL: srv.URL + "/a1/", Title: "A1"},
		{URL: srv.URL + "/a2/", Title: "A2"},
		{URL: srv.URL + "/a3/", Title: "A3"},
	}
	cfg := SourceConfig{MaxItemsPerRun: 2, TimeoutSec: 5}

	res := &RunResult{}
	importListQueue(db, files, "T", "Q", cfg, queue, res)
	if res.Imported != 2 {
		t.Fatalf("first run imported=%d skipped=%d want imported=2", res.Imported, res.Skipped)
	}

	res2 := &RunResult{}
	importListQueue(db, files, "T", "Q", cfg, queue, res2)
	if res2.Skipped != 2 || res2.Imported != 1 {
		t.Fatalf("second run skipped=%d imported=%d want 2/1 (page-2 backfill)", res2.Skipped, res2.Imported)
	}
}
