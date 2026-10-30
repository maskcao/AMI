package pipeline

import "testing"

func TestExtractDisclosedAtFromHTML_meta(t *testing.T) {
	html := `<html><head><meta property="article:published_time" content="2025-07-30T03:25:58.000Z"></head></html>`
	got := ExtractDisclosedAtFromHTML(html)
	want := "2025-07-30T03:25:58Z"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExtractDisclosedAtFromHTML_nuxt(t *testing.T) {
	html := `window.__NUXT__=function(e){return {publish_time="2025-06-24T01:52:14.000Z"}}`
	got := ExtractDisclosedAtFromHTML(html)
	want := "2025-06-24T01:52:14Z"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
