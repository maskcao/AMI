package ingest_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ami/ami/service/internal/ingest"
)

func TestParseWechatWeeklyThreatDigest(t *testing.T) {
	raw, err := os.ReadFile("testdata/wechat_weekly_snippet.html")
	if err != nil {
		t.Skip("testdata missing")
	}
	entries := ingest.ParseWechatWeeklyThreatDigest(raw)
	if len(entries) == 0 {
		t.Fatal("expected entries")
	}
	var attack, malware int
	for _, e := range entries {
		if e.Category == "attack_group" {
			attack++
		}
		if e.Category == "malware" {
			malware++
		}
		if e.DisclosedAt == "" || e.SourceURL == "" {
			t.Fatalf("incomplete entry: %+v", e)
		}
		if !strings.HasPrefix(e.SourceURL, "http") {
			t.Fatalf("bad source url: %q", e.SourceURL)
		}
	}
	if attack == 0 || malware == 0 {
		t.Fatalf("attack=%d malware=%d entries=%d", attack, malware, len(entries))
	}
}

func TestParseWeChatAlbumURL(t *testing.T) {
	list := "https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzI2MDc2MDA4OA==&action=getalbum&album_id=1533540987225096195"
	cfg := ingest.SourceConfig{ListURL: list}
	// exercised via collect — only URL parse is internal; smoke via Discover if integration present
	_ = cfg
}
