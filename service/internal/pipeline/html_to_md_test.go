package pipeline

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdownHeadings(t *testing.T) {
	html := `<html><body><h1>Title</h1><p>Hello <b>world</b>.</p></body></html>`
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "# Title") || !strings.Contains(md, "Hello") || !strings.Contains(md, "world") {
		t.Fatalf("unexpected md: %q", md)
	}
}
