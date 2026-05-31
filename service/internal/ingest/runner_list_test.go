package ingest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/store"
)

func TestRunPaginatedListCollect(t *testing.T) {
	article := `<html><head><title>Bot Article</title></head><body><p>CVE-2099-0001</p></body></html>`
	list2 := `<html></html>`

	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tag/botnet/":
			list1 := fmt.Sprintf(`<a href="%s/2024/a1/">A1</a><a href="/tag/x/">skip</a>`, base)
			_, _ = w.Write([]byte(list1))
		case "/tag/botnet/page/2/":
			_, _ = w.Write([]byte(list2))
		case "/2024/a1/":
			_, _ = w.Write([]byte(article))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	base = strings.TrimRight(srv.URL, "/")
	host, _ := url.Parse(srv.URL)

	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(dir)
	ts := time.Now().UTC().Format(time.RFC3339)
	cfg := fmt.Sprintf(`{
		"list_url":"%s/tag/botnet/",
		"pagination":{"type":"path","path_template":"%s/tag/botnet/page/{page}/","start_page":2,"max_pages":1,"stop_on_empty_list":true},
		"link_extract":{"must_contain":"%s","exclude_contains":["/tag/"],"prefer_path_regex":"/20[0-9]{2}/"},
		"article_base":"%s",
		"max_items_per_run":5
	}`, base, base, host.Host, base)
	_, err = db.Exec(`INSERT INTO collect_source (id, name, type, enabled, schedule_cron, schedule_label, source_label, config_json, updated_at)
		VALUES ('cs-list', 'Test List', 'paginated_list', 1, '0 * * * *', 'hourly', 'TEST', ?, ?)`, cfg, ts)
	if err != nil {
		t.Fatal(err)
	}

	res, err := RunCollectSource(db, files, "cs-list")
	if err != nil {
		t.Fatal(err)
	}
	if res.Imported != 1 {
		t.Fatalf("imported=%d skipped=%d failed=%d", res.Imported, res.Skipped, res.Failed)
	}
}
