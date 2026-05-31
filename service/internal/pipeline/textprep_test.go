package pipeline

import (
	"os"
	"strings"
	"testing"
)

func TestPrepareArticleTextStripsSchemaOrg(t *testing.T) {
	html := `<html><head><script type="application/ld+json">{"@context":"https://schema.org","@type":"Article"}</script></head>
<body><h1>ZHtrap botnet</h1><p>Uses Mirai code against IoT devices.</p></body></html>`
	text := PrepareArticleText(html)
	if strings.Contains(text, `"@context"`) {
		t.Fatalf("schema.org leaked: %q", text)
	}
	if !strings.Contains(text, "ZHtrap") {
		t.Fatalf("missing body: %q", text)
	}
}

func TestPrepareArticleTextNetlabSample(t *testing.T) {
	b, err := os.ReadFile("../../data/store/intel/52/raw/001.html")
	if err != nil {
		t.Skip("no sample html")
	}
	text := PrepareArticleText(string(b))
	if strings.Contains(text, `"@context": "https://schema.org"`) {
		t.Fatalf("schema.org still present, len=%d head=%q", len(text), text[:min(200, len(text))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
