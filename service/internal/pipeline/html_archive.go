package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/netutil"
)

var (
	linkCSSRe    = regexp.MustCompile(`(?is)<link\b[^>]*\brel=["']stylesheet["'][^>]*\bhref=["']([^"']+)["']`)
	linkCSSRe2   = regexp.MustCompile(`(?is)<link\b[^>]*\bhref=["']([^"']+)["'][^>]*\brel=["']stylesheet["']`)
	scriptSrcRe  = regexp.MustCompile(`(?is)<script\b[^>]*\bsrc=["']([^"']+)["']`)
	imgDataSrcRe = regexp.MustCompile(`(?is)<img\b[^>]*\bdata-(?:src|original|lazy-src)=["']([^"']+)["']`)
	imgSrcsetRe  = regexp.MustCompile(`(?is)<img\b[^>]*\bsrcset=["']([^"']+)["']`)
	// Qianxin/TI articles embed images as site-relative /uploads/... paths in SSR and NUXT content.
	uploadsImagePathRe = regexp.MustCompile(`(?i)/uploads/\d{4}/\d{2}/\d{2}/[a-f0-9]+\.(?:png|jpe?g|webp|gif)`)
)

const (
	maxPageImages      = 120
	maxPageOtherAssets = 30
)

// ArchivePageAssets downloads linked images (priority) plus CSS/JS and rewrites HTML to local API paths.
func ArchivePageAssets(files *artifacts.Store, intelID int64, pageURL string, html []byte) ([]byte, int, error) {
	if files == nil || len(html) == 0 {
		return html, 0, nil
	}
	if os.Getenv("AMI_SKIP_ASSET_ARCHIVE") == "1" {
		return html, 0, nil
	}
	base, _ := url.Parse(pageURL)
	client := &http.Client{Timeout: 20 * time.Second}
	seenAbs := map[string]string{}
	rawToLocal := map[string]string{}
	imgCount := 0
	otherCount := 0
	out := string(html)

	archiveRef := func(raw string, image bool) string {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "data:") || strings.HasPrefix(raw, "#") {
			return raw
		}
		if strings.HasPrefix(raw, "/api/v1/intel/") {
			return raw
		}
		if local, ok := rawToLocal[raw]; ok {
			return local
		}
		abs := resolveFigureURL(base, raw)
		if abs == "" {
			return raw
		}
		if local, ok := seenAbs[abs]; ok {
			rawToLocal[raw] = local
			return local
		}
		if image {
			if imgCount >= maxPageImages {
				return raw
			}
		} else if otherCount >= maxPageOtherAssets {
			return raw
		}
		body, mime, err := fetchAsset(client, abs, pageURL)
		if err != nil || len(body) == 0 {
			return raw
		}
		if image && !IsImageBytes(body) {
			return raw
		}
		ext := extForMime(mime, abs)
		sum := sha256.Sum256([]byte(abs))
		name := hex.EncodeToString(sum[:8]) + ext
		if _, err := files.WriteRawAsset(intelID, 1, name, body); err != nil {
			return raw
		}
		local := fmt.Sprintf("/api/v1/intel/%d/raw-assets/%s", intelID, name)
		seenAbs[abs] = local
		rawToLocal[raw] = local
		if image {
			imgCount++
		} else {
			otherCount++
		}
		return local
	}

	// Pass 1: collect and archive all image candidates (img tags, NUXT /uploads/ paths, etc.).
	for _, raw := range enumerateImageRefs(out) {
		archiveRef(raw, true)
	}
	out = applyURLRewrites(out, rawToLocal)

	replaceAttr := func(re *regexp.Regexp, image bool, pick func(string) string) {
		out = re.ReplaceAllStringFunc(out, func(tag string) string {
			sub := re.FindStringSubmatch(tag)
			if len(sub) < 2 {
				return tag
			}
			raw := sub[1]
			if pick != nil {
				raw = pick(raw)
			}
			local := archiveRef(raw, image)
			if local == sub[1] {
				return tag
			}
			return strings.Replace(tag, sub[1], local, 1)
		})
	}

	replaceAttr(imgSrcRe, true, nil)
	replaceAttr(imgDataSrcRe, true, nil)
	replaceAttr(imgSrcsetRe, true, bestSrcsetURL)

	// Pass 2: any /uploads/... still embedded in HTML strings (e.g. inside __NUXT__ content).
	for _, raw := range uploadsImagePathRe.FindAllString(out, -1) {
		local := archiveRef(raw, true)
		if local != raw {
			out = strings.ReplaceAll(out, raw, local)
		}
	}

	for _, re := range []*regexp.Regexp{linkCSSRe, linkCSSRe2, scriptSrcRe} {
		replaceAttr(re, false, nil)
	}
	out = HydrateLazyLoadImages(out)
	return []byte(out), imgCount + otherCount, nil
}

func enumerateImageRefs(html string) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "data:") || strings.HasPrefix(raw, "#") {
			return
		}
		if strings.HasPrefix(raw, "/api/v1/intel/") {
			return
		}
		if _, ok := seen[raw]; ok {
			return
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	for _, re := range []*regexp.Regexp{imgSrcRe, imgDataSrcRe} {
		for _, m := range re.FindAllStringSubmatch(html, -1) {
			if len(m) > 1 {
				add(m[1])
			}
		}
	}
	for _, m := range imgSrcsetRe.FindAllStringSubmatch(html, -1) {
		if len(m) > 1 {
			add(bestSrcsetURL(m[1]))
		}
	}
	for _, path := range uploadsImagePathRe.FindAllString(html, -1) {
		add(path)
	}
	return out
}

func applyURLRewrites(html string, rawToLocal map[string]string) string {
	if len(rawToLocal) == 0 {
		return html
	}
	keys := make([]string, 0, len(rawToLocal))
	for raw, local := range rawToLocal {
		if raw != local {
			keys = append(keys, raw)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, raw := range keys {
		html = strings.ReplaceAll(html, raw, rawToLocal[raw])
	}
	return html
}

func bestSrcsetURL(srcset string) string {
	srcset = strings.TrimSpace(srcset)
	if srcset == "" {
		return ""
	}
	parts := strings.Split(srcset, ",")
	best := ""
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		fields := strings.Fields(p)
		u := fields[0]
		if best == "" {
			best = u
			continue
		}
		if len(fields) > 1 && strings.HasSuffix(fields[1], "w") {
			best = u
		}
	}
	return best
}

func fetchAsset(client *http.Client, assetURL, pageURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", netutil.DefaultUserAgent)
	if pageURL != "" {
		req.Header.Set("Referer", pageURL)
	}
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

// isNonArticleImageRef skips theme/nav chrome when auditing remote images (e.g. Unit42 tag pages).
func isNonArticleImageRef(ref string) bool {
	lower := strings.ToLower(ref)
	if strings.Contains(lower, "/wp-content/themes/") {
		return true
	}
	if strings.Contains(lower, "/dist/images/icons/") {
		return true
	}
	if strings.Contains(lower, "gravatar.com") {
		return true
	}
	if strings.Contains(lower, "/content/dam/") || strings.Contains(lower, "/etc/clientlibs/") {
		return true
	}
	if strings.HasSuffix(lower, ".svg") {
		return true
	}
	if strings.Contains(lower, "footer-logo") || strings.Contains(lower, "top-threats.svg") {
		return true
	}
	return false
}

// AuditHTMLImages reports unresolved image refs in archived HTML (for batch checks).
func AuditHTMLImages(html []byte) (total, local, unresolvedUploads, remote int) {
	s := string(html)
	if i := strings.Index(s, "window.__NUXT__"); i > 0 {
		s = s[:i]
	}
	srcs := imgSrcRe.FindAllStringSubmatch(s, -1)
	total = len(srcs)
	for _, m := range srcs {
		if len(m) < 2 {
			continue
		}
		ref := m[1]
		if isNonArticleImageRef(ref) {
			continue
		}
		switch {
		case strings.Contains(ref, "/raw-assets/"):
			local++
		case strings.HasPrefix(ref, "/uploads/"):
			unresolvedUploads++
		case strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "//"):
			remote++
		}
	}
	return total, local, unresolvedUploads, remote
}
