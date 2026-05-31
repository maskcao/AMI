package ingest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/store"
)

func TestRunRSSCollect(t *testing.T) {
	article := `<html><body><p>Exploit CVE-2099-0001 in the wild.</p></body></html>`
	var feed string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed.xml":
			_, _ = w.Write([]byte(feed))
		case "/article/1":
			_, _ = w.Write([]byte(article))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	feed = `<?xml version="1.0"?><rss><channel>
<item><title>Test Advisory</title><link>` + srv.URL + `/article/1</link></item>
</channel></rss>`

	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(dir)
	ts := time.Now().UTC().Format(time.RFC3339)
	cfg := `{"feed_url":"` + srv.URL + `/feed.xml","max_items_per_run":3}`
	_, err = db.Exec(`INSERT INTO collect_source (id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, updated_at)
		VALUES ('cs-test', 'Test RSS', 'rss', 1, '0 * * * *', 'hourly', 'TEST', ?, ?)`, cfg, ts)
	if err != nil {
		t.Fatal(err)
	}

	res, err := RunCollectSource(db, files, "cs-test")
	if err != nil {
		t.Fatal(err)
	}
	if res.Imported != 1 {
		t.Fatalf("imported=%d skipped=%d failed=%d", res.Imported, res.Skipped, res.Failed)
	}
	exists, _ := db.IntelExistsBySourceURL(srv.URL + "/article/1")
	if !exists {
		t.Fatal("expected intel row")
	}
}
