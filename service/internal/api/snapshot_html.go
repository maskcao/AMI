package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/ami/ami/service/internal/pipeline"
)

var snapshotBaseTagRe = regexp.MustCompile(`(?is)<base\b[^>]*\/?>`)

// stripSnapshotBase removes <base href="..."> from archived portal HTML. Injected via v-html it can
// hijack relative URL resolution for the whole SPA (e.g. Qianxin /blog/ base breaks asset paths).
func stripSnapshotBase(html string) string {
	return snapshotBaseTagRe.ReplaceAllString(html, "")
}

func requestPublicOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); p != "" {
		scheme = strings.ToLower(strings.Split(p, ",")[0])
	}
	// Vite dev proxy uses changeOrigin; prefer X-Forwarded-Host so assets use :5173 not 127.0.0.1:8080.
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		host = "127.0.0.1:8080"
	}
	return scheme + "://" + strings.TrimSuffix(host, "/")
}

func rewriteSnapshotAssetOrigins(html string, origin string, intelID int64) string {
	if origin == "" || intelID <= 0 {
		return html
	}
	origin = strings.TrimSuffix(origin, "/")
	for _, segment := range []string{"raw-assets", "artifacts"} {
		prefix := fmt.Sprintf("/api/v1/intel/%d/%s/", intelID, segment)
		html = strings.ReplaceAll(html, prefix, origin+prefix)
	}
	return html
}

func prepareSnapshotHTML(html []byte, r *http.Request, intelID int64) []byte {
	_ = r
	_ = intelID
	// Serve relative /api/v1/intel/... asset paths; WebUI adds window.location.origin (avoids double-rewrite).
	s := stripSnapshotBase(string(html))
	s = pipeline.HydrateLazyLoadImages(s)
	return []byte(s)
}
