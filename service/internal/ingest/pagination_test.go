package ingest

import "testing"

func TestExpandListPageURLs_path(t *testing.T) {
	cfg := SourceConfig{
		ListURL: "https://example.com/tag/botnet/",
		Pagination: PaginationCfg{
			Type:         "path",
			PathTemplate: "https://example.com/tag/botnet/page/{page}/",
			StartPage:    2,
			MaxPages:     2,
		},
	}
	got := ExpandListPageURLs(cfg)
	if len(got) != 3 {
		t.Fatalf("len=%d urls=%v", len(got), got)
	}
	if got[0] != cfg.ListURL {
		t.Fatalf("first=%s", got[0])
	}
	if got[1] != "https://example.com/tag/botnet/page/2/" {
		t.Fatalf("page2=%s", got[1])
	}
}

func TestExtractArticleLinks(t *testing.T) {
	html := []byte(`
<html><body>
<a href="/tag/botnet/">tag</a>
<a href="https://blog.netlab.360.com/2024/foo/">Article One</a>
<a href="https://blog.netlab.360.com/author/x/">author</a>
</body></html>`)
	cfg := SourceConfig{
		LinkExtract: LinkExtractCfg{
			MustContain:     "blog.netlab.360.com",
			ExcludeContains: []string{"/tag/", "/author/"},
			PreferPathRegex: "/20[0-9]{2}/",
		},
	}
	links, err := extractArticleLinks(html, "https://blog.netlab.360.com/", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].URL != "https://blog.netlab.360.com/2024/foo/" {
		t.Fatalf("links=%+v", links)
	}
}
