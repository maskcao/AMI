package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
)

var (
	linkCSSRe  = regexp.MustCompile(`(?is)<link\b[^>]*\brel=["']stylesheet["'][^>]*\bhref=["']([^"']+)["']`)
	linkCSSRe2 = regexp.MustCompile(`(?is)<link\b[^>]*\bhref=["']([^"']+)["'][^>]*\brel=["']stylesheet["']`)
	scriptSrcRe = regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=["']([^"']+)["']`)
)

const maxPageAssets = 40

// ArchivePageAssets downloads linked CSS/JS/images and rewrites HTML to local API paths.
func ArchivePageAssets(files *artifacts.Store, intelID int64, pageURL string, html []byte) ([]byte, int, error) {
	if files == nil || len(html) == 0 {
		return html, 0, nil
	}
	base, _ := url.Parse(pageURL)
	client := &http.Client{Timeout: 20 * time.Second}
	seen := map[string]string{}
	count := 0
	out := string(html)

	replaceRef := func(raw string) string {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "data:") || strings.HasPrefix(raw, "#") {
			return raw
		}
		abs := resolveFigureURL(base, raw)
		if abs == "" {
			return raw
		}
		if local, ok := seen[abs]; ok {
			return local
		}
		if count >= maxPageAssets {
			return raw
		}
		body, mime, err := fetchAsset(client, abs)
		if err != nil || len(body) == 0 {
			return raw
		}
		ext := extForMime(mime, abs)
		sum := sha256.Sum256([]byte(abs))
		name := hex.EncodeToString(sum[:8]) + ext
		if _, err := files.WriteRawAsset(intelID, 1, name, body); err != nil {
			return raw
		}
		local := fmt.Sprintf("/api/v1/intel/%d/raw-assets/%s", intelID, name)
		seen[abs] = local
		count++
		return local
	}

	for _, re := range []*regexp.Regexp{imgSrcRe, linkCSSRe, linkCSSRe2, scriptSrcRe} {
		out = re.ReplaceAllStringFunc(out, func(tag string) string {
			sub := re.FindStringSubmatch(tag)
			if len(sub) < 2 {
				return tag
			}
			local := replaceRef(sub[1])
			if local == sub[1] {
				return tag
			}
			return strings.Replace(tag, sub[1], local, 1)
		})
	}
	return []byte(out), count, nil
}

func fetchAsset(client *http.Client, assetURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "AMI-Ingest/1.0")
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, "", fmt.Errorf("http %d", res.StatusCode)
	}
	const maxBytes = 5 << 20
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBytes))
	if err != nil {
		return nil, "", err
	}
	mime := res.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	if i := strings.Index(mime, ";"); i > 0 {
		mime = mime[:i]
	}
	return body, mime, nil
}
