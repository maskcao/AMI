package ingest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverWordPressAjaxLinks(t *testing.T) {
	ajaxHits := 0
	var baseURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tag/botnet/" {
			body := fmt.Sprintf(`<!doctype html><html><body>
<div id="loadMoreblog" data-load-posts data-tracking="post_tag:botnet" data-lang="en"></div>
<script>
var blog_loadmore_params = {"ajaxurl":"%s/wp-admin/admin-ajax.php","posts":"{\"tag\":\"botnet\"}","current_page":"1","max_page":"2","nonce":"test-nonce"};
</script></body></html>`, baseURL)
			w.Write([]byte(body))
			return
		}
		if r.URL.Path == "/wp-admin/admin-ajax.php" {
			_ = r.ParseForm()
			ajaxHits++
			page := r.FormValue("page")
			if page == "1" {
				w.Write([]byte(`{"html":"<a href=\"https://example.com/unit42-article-one/\">One</a>"}`))
				return
			}
			w.Write([]byte(`{"html":"<a href=\"https://example.com/unit42-article-two/\">Two</a>"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	baseURL = srv.URL

	cfg := SourceConfig{
		ListURL:     srv.URL + "/tag/botnet/",
		ArticleBase: srv.URL,
		LinkExtract: LinkExtractCfg{
			MustContain:     "example.com",
			ExcludeContains: []string{"/tag/"},
		},
		Pagination: PaginationCfg{
			Type:            "wordpress_ajax",
			MaxPages:        2,
			StopOnEmptyList: true,
		},
	}
	disc, err := discoverWordPressAjaxLinks(cfg, 0)
	if err != nil {
		t.Fatal(err)
	}
	if disc.LinkCount < 2 {
		t.Fatalf("link count=%d want >=2", disc.LinkCount)
	}
	if ajaxHits < 2 {
		t.Fatalf("ajax hits=%d want >=2", ajaxHits)
	}
}
