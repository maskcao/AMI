package ingest

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/netutil"
)

func fetchArticleHTML(rawURL string, timeoutSec int) ([]byte, error) {
	return httpGet(rawURL, timeoutSec)
}

// FetchArticleHTML downloads a page body for repair/re-archive jobs.
func FetchArticleHTML(rawURL string, timeoutSec int) ([]byte, error) {
	return fetchArticleHTML(rawURL, timeoutSec)
}

func resolveURL(base *url.URL, href string) string {
	if base == nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref).String()
	if strings.HasPrefix(strings.ToLower(abs), "http") {
		return abs
	}
	return ""
}

func httpGet(url string, timeoutSec int) ([]byte, error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	ua := netutil.DefaultUserAgent
	if strings.Contains(strings.ToLower(url), "mp.weixin.qq.com") {
		ua = netutil.WeChatUserAgent
	}
	req.Header.Set("User-Agent", ua)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: %d", url, res.StatusCode)
	}
	const maxBody = 8 << 20
	return readLimitedBody(res.Body, maxBody)
}

func readLimitedBody(r io.Reader, max int64) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, max))
}
