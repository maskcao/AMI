package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_prepareSnapshotHTML_stripsBaseAndRewrites(t *testing.T) {
	raw := `<html><head><base href="/blog/"></head><body><img src="/api/v1/intel/179/raw-assets/ab.png"></body></html>`
	r := httptest.NewRequest("GET", "http://192.168.238.10:5173/api/v1/intel/179/original", nil)
	r.Host = "127.0.0.1:8080"
	r.Header.Set("X-Forwarded-Host", "192.168.238.10:5173")
	r.Header.Set("X-Forwarded-Proto", "http")
	out := string(prepareSnapshotHTML([]byte(raw), r, 179))
	if strings.Contains(strings.ToLower(out), "<base") {
		t.Fatalf("base tag not stripped: %s", out)
	}
	want := "/api/v1/intel/179/raw-assets/ab.png"
	if !strings.Contains(out, want) {
		t.Fatalf("got %q want contains %q", out, want)
	}
	if strings.Contains(strings.ToLower(out), "<base") {
		t.Fatalf("base tag not stripped")
	}
	if strings.Contains(out, "http://192.168.238.10:5173") {
		t.Fatalf("server should not inject origin: %s", out)
	}
}

func Test_prepareSnapshotHTML_hydratesLozad(t *testing.T) {
	raw := `<img class="lozad" data-src="/api/v1/intel/237/raw-assets/x.png" alt="fig">`
	r := httptest.NewRequest("GET", "/api/v1/intel/237/original", nil)
	out := string(prepareSnapshotHTML([]byte(raw), r, 237))
	if !strings.Contains(out, `src="/api/v1/intel/237/raw-assets/x.png"`) {
		t.Fatalf("lozad not hydrated: %s", out)
	}
}
