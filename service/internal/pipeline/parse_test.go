package pipeline

import "testing"

func TestExtractTitleFromHTML(t *testing.T) {
	html := `<html><head>
<meta property="og:title" content="Botnet Orchard Advisory" />
<title>ignored title</title></head>
<body><h1>H1 fallback</h1></body></html>`
	got := ExtractTitleFromHTML(html)
	if got != "Botnet Orchard Advisory" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractTitleFromHTML_ghostHeadline(t *testing.T) {
	html := `<script type="application/ld+json">{"@type":"Article","headline":"ZHtrap僵尸网络分析报告"}</script>
<title>ZHtrap僵尸网络分析报告 - 360 Netlab Blog</title>
<body><h1 class="post-full-title">ZHtrap僵尸网络分析报告</h1>
<p>版权 版权声明 本文为Netlab原创 很长正文……</p></body>`
	got := ExtractTitleFromHTML(html)
	if got != "ZHtrap僵尸网络分析报告" {
		t.Fatalf("got %q", got)
	}
}

func TestIsPlausibleTitle_rejectsListBlurb(t *testing.T) {
	blurb := "New Threat 新威胁: ZHtrap僵尸网络分析报告 版权 版权声明: 本文为Netlab原创，转载请标明来源360Netlab。背景介绍 2021年12月……"
	if IsPlausibleTitle(blurb) {
		t.Fatal("expected list blurb rejected")
	}
	if !IsPlausibleTitle("ZHtrap僵尸网络分析报告") {
		t.Fatal("expected short title accepted")
	}
}

func TestPickIngestTitle_prefersHTML(t *testing.T) {
	html := []byte(`<html><head><meta property="og:title" content="ZHtrap僵尸网络分析报告" /></head></html>`)
	blurb := "New Threat 新威胁: ZHtrap僵尸网络分析报告 版权 版权声明 本文为Netlab原创……"
	got := PickIngestTitle(html, "https://blog.netlab.360.com/zhtrap-report/", blurb)
	if got != "ZHtrap僵尸网络分析报告" {
		t.Fatalf("got %q", got)
	}
}

func TestShouldRefreshIntelTitle(t *testing.T) {
	if !ShouldRefreshIntelTitle("", "https://x/a", "Real Title") {
		t.Fatal("empty current")
	}
	if !ShouldRefreshIntelTitle("https://x/a", "https://x/a", "Real Title") {
		t.Fatal("url as title")
	}
	long := string(make([]rune, 200)) + "版权所有"
	if !ShouldRefreshIntelTitle(long, "https://x/a", "Real Title") {
		t.Fatal("long bad title should refresh")
	}
	if ShouldRefreshIntelTitle("Stable Report Title", "https://x/a", "Other Title") {
		t.Fatal("should keep good title")
	}
}
