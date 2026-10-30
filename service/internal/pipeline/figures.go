package pipeline

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/netutil"
	"github.com/ami/ami/service/internal/store"
)

var imgSrcRe = regexp.MustCompile(`(?is)<img\b[^>]*\bsrc=["']([^"']+)["']`)

// SyncFiguresFromHTML downloads report images and registers intel_figure rows.
func SyncFiguresFromHTML(files *artifacts.Store, db *store.DB, intelID int64, pageURL string, html []byte) error {
	if files == nil || db == nil || len(html) == 0 {
		return nil
	}
	if os.Getenv("AMI_SKIP_ASSET_ARCHIVE") == "1" {
		return nil
	}
	base, _ := url.Parse(pageURL)
	matches := imgSrcRe.FindAllStringSubmatch(string(html), -1)
	if len(matches) == 0 {
		return nil
	}
	client := &http.Client{Timeout: 25 * time.Second}
	seen := make(map[string]struct{})
	idx := 0
	for _, m := range matches {
		if idx >= 12 {
			break
		}
		rawSrc := strings.TrimSpace(m[1])
		if rawSrc == "" || strings.HasPrefix(rawSrc, "data:") {
			continue
		}
		cacheKey := rawSrc
		var body []byte
		var mime string
		var err error
		if name := rawAssetName(intelID, rawSrc); name != "" {
			cacheKey = "local:" + name
			body, err = files.ReadRawAsset(intelID, 1, name)
			if err != nil || len(body) < 2048 {
				continue
			}
			mime = mimeFromExt(filepath.Ext(name))
		} else {
			abs := resolveFigureURL(base, rawSrc)
			if abs == "" {
				continue
			}
			cacheKey = abs
			if isDecorativeImage(abs) {
				continue
			}
			body, mime, err = fetchImage(client, abs)
			if err != nil || len(body) < 2048 {
				continue
			}
		}
		if _, ok := seen[cacheKey]; ok {
			continue
		}
		seen[cacheKey] = struct{}{}
		ext := extForMime(mime, rawSrc)
		figID := fmt.Sprintf("fig-%03d", idx+1)
		rel, err := files.WriteFigure(intelID, "v1", figID+ext, body)
		if err != nil {
			continue
		}
		apiSrc := fmt.Sprintf("/api/v1/intel/%d/artifacts/%s", intelID, path.Base(rel))
		title := figureTitleFromURL(rawSrc, idx)
		_ = db.InsertFigure(store.IntelFigure{
			ID:      figID,
			IntelID: intelID,
			Title:   title,
			Caption: title,
			Kind:    "screenshot",
			Src:     apiSrc,
		})
		idx++
	}
	return nil
}

func resolveFigureURL(base *url.URL, raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if base != nil && !u.IsAbs() {
		u = base.ResolveReference(u)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return u.String()
}

func isDecorativeImage(u string) bool {
	low := strings.ToLower(u)
	for _, frag := range []string{
		"favicon", "avatar", "logo", "icon", "badge", "gravatar",
		"/size/w30/", "/size/w60/", "/author/", "netlab-brand", "contributor",
	} {
		if strings.Contains(low, frag) {
			return true
		}
	}
	return false
}

func fetchImage(client *http.Client, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", netutil.DefaultUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, "", fmt.Errorf("http %d", res.StatusCode)
	}
	const maxBytes = 6 << 20
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBytes))
	if err != nil {
		return nil, "", err
	}
	mime := res.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/png"
	}
	if i := strings.Index(mime, ";"); i > 0 {
		mime = mime[:i]
	}
	return body, mime, nil
}

func extForMime(mime, src string) string {
	switch mime {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	default:
		if ext := path.Ext(src); ext != "" && len(ext) <= 5 {
			return ext
		}
		return ".png"
	}
}

func rawAssetName(intelID int64, raw string) string {
	raw = strings.Split(raw, "?")[0]
	prefix := fmt.Sprintf("/api/v1/intel/%d/raw-assets/", intelID)
	if i := strings.Index(raw, prefix); i >= 0 {
		return filepath.Base(raw[i+len(prefix):])
	}
	if strings.HasPrefix(raw, "/raw-assets/") {
		return filepath.Base(raw)
	}
	return ""
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "image/png"
	}
}

func figureTitleFromURL(imgURL string, idx int) string {
	if u, err := url.Parse(imgURL); err == nil {
		base := path.Base(u.Path)
		if base != "" && base != "/" && base != "." {
			return base
		}
	}
	return fmt.Sprintf("报告插图 %d", idx+1)
}
