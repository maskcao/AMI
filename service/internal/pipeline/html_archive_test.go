package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ami/ami/service/internal/artifacts"
)

func TestEnumerateImageRefs_uploads(t *testing.T) {
	html := `<img src="/uploads/2025/12/02/abc.png"><p>/uploads/2025/12/02/def.jpg in text`
	refs := enumerateImageRefs(html)
	if len(refs) < 2 {
		t.Fatalf("refs=%v want >=2", refs)
	}
}

func TestAuditHTMLImages_skipsThemeChrome(t *testing.T) {
	html := []byte(`<img src="https://unit42.paloaltonetworks.com/wp-content/themes/unit42-v6/dist/images/icons/icon-clock.svg">
<img src="https://unit42.paloaltonetworks.com/wp-content/uploads/2026/04/body.png">`)
	total, _, _, remote := AuditHTMLImages(html)
	if total != 2 {
		t.Fatalf("total=%d want 2", total)
	}
	if remote != 1 {
		t.Fatalf("remote=%d want 1 (uploads body image only)", remote)
	}
}

func TestAuditHTMLImages_unresolvedUploads(t *testing.T) {
	html := []byte(`<img src="/uploads/2025/12/02/abc.png">`)
	total, local, unresolved, remote := AuditHTMLImages(html)
	if total != 1 || local != 0 || unresolved != 1 || remote != 0 {
		t.Fatalf("got total=%d local=%d unresolved=%d remote=%d", total, local, unresolved, remote)
	}
}

func TestArchivePageAssets_rewritesUploads(t *testing.T) {
	dir := t.TempDir()
	files := artifacts.New(dir)
	html := []byte(`<html><body><img src="/uploads/2025/12/02/1e8bb76b973e87607a26d27a12b33929.png"></body></html>`)
	page := "https://ti.qianxin.com/blog/articles/sample-en"
	out, n, err := ArchivePageAssets(files, 99, page, html)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("expected assets archived")
	}
	s := string(out)
	if strings.Contains(s, "/uploads/2025/12/02/") {
		t.Fatalf("uploads path not rewritten: %s", s)
	}
	if !strings.Contains(s, "/api/v1/intel/99/raw-assets/") {
		t.Fatalf("expected raw-assets path in %s", s)
	}
	assetsDir := filepath.Join(dir, "store", "intel", "99", "raw", "001", "assets")
	entries, _ := os.ReadDir(assetsDir)
	if len(entries) == 0 {
		t.Fatal("no assets on disk")
	}
}

