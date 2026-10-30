package ingest

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/netutil"
)

// collectWeChatAlbumLinks pages mp.weixin.qq.com album JSON API (scroll/load-more equivalent).
func collectWeChatAlbumLinks(cfg SourceConfig, maxPages int) ([]listLink, int, error) {
	biz, albumID, err := parseWeChatAlbumURL(cfg.ListURL)
	if err != nil {
		return nil, 0, err
	}
	if maxPages <= 0 {
		maxPages = cfg.Pagination.MaxPages
	}
	if maxPages <= 0 {
		maxPages = 50
	}
	perPage := cfg.Pagination.ScrollSteps
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}
	isReverse := 1
	if cfg.Pagination.RenderMode == "asc" {
		isReverse = 0
	}

	titleNeedle := strings.TrimSpace(cfg.LinkExtract.TitleMustContain)
	if titleNeedle == "" {
		titleNeedle = "每周高级威胁情报解读"
	}

	client := &http.Client{Timeout: time.Duration(cfg.TimeoutSec) * time.Second}
	if client.Timeout <= 0 {
		client.Timeout = 45 * time.Second
	}

	seen := map[string]struct{}{}
	var out []listLink
	pages := 0
	beginMsgID := ""
	beginItemIdx := 1

	for pages < maxPages {
		body, err := fetchWeChatAlbumPage(client, biz, albumID, perPage, isReverse, beginMsgID, beginItemIdx)
		if err != nil {
			return out, pages, err
		}
		var resp wechatAlbumJSON
		if err := json.Unmarshal(body, &resp); err != nil {
			return out, pages, fmt.Errorf("wechat album json: %w", err)
		}
		if resp.BaseResp.Ret != 0 {
			return out, pages, fmt.Errorf("wechat album api ret=%d", resp.BaseResp.Ret)
		}
		articles := resp.GetAlbumResp.ArticleList
		if len(articles) == 0 {
			break
		}
		pages++
		for _, a := range articles {
			title := strings.TrimSpace(a.Title)
			if titleNeedle != "" && !strings.Contains(title, titleNeedle) {
				continue
			}
			rawURL := html.UnescapeString(strings.TrimSpace(a.URL))
			if rawURL == "" {
				continue
			}
			if _, ok := seen[rawURL]; ok {
				continue
			}
			seen[rawURL] = struct{}{}
			out = append(out, listLink{URL: rawURL, Title: title})
		}
		last := articles[len(articles)-1]
		beginMsgID = strings.TrimSpace(last.MsgID)
		if beginMsgID == "" {
			break
		}
		if idx, err := strconv.Atoi(strings.TrimSpace(last.ItemIdx)); err == nil && idx > 0 {
			beginItemIdx = idx
		} else {
			beginItemIdx = 1
		}
		if int(resp.GetAlbumResp.ContinueFlag) == 0 {
			break
		}
		sleepDelay(cfg.DelayMS)
	}
	return out, pages, nil
}

func parseWeChatAlbumURL(listURL string) (biz, albumID string, err error) {
	u, err := url.Parse(strings.TrimSpace(listURL))
	if err != nil {
		return "", "", err
	}
	q := u.Query()
	biz = strings.TrimSpace(q.Get("__biz"))
	albumID = strings.TrimSpace(q.Get("album_id"))
	if biz == "" || albumID == "" {
		return "", "", fmt.Errorf("list_url must include __biz and album_id query params")
	}
	return biz, albumID, nil
}

func fetchWeChatAlbumPage(client *http.Client, biz, albumID string, count, isReverse int, beginMsgID string, beginItemIdx int) ([]byte, error) {
	q := url.Values{}
	q.Set("action", "getalbum")
	q.Set("__biz", biz)
	q.Set("album_id", albumID)
	q.Set("count", strconv.Itoa(count))
	q.Set("is_reverse", strconv.Itoa(isReverse))
	q.Set("f", "json")
	q.Set("x5", "0")
	if beginMsgID != "" {
		q.Set("begin_msgid", beginMsgID)
		q.Set("begin_itemidx", strconv.Itoa(beginItemIdx))
	}
	endpoint := "https://mp.weixin.qq.com/mp/appmsgalbum?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", netutil.WeChatUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("GET album: %d", res.StatusCode)
	}
	return readLimitedBody(res.Body, 8<<20)
}

type wechatAlbumJSON struct {
	BaseResp struct {
		Ret int `json:"ret"`
	} `json:"base_resp"`
	GetAlbumResp struct {
		ArticleList  []wechatAlbumArticle `json:"article_list"`
		ContinueFlag flexInt              `json:"continue_flag"`
	} `json:"getalbum_resp"`
}

// flexInt accepts JSON number or numeric string from WeChat APIs.
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		*f = 0
		return nil
	}
	*f = flexInt(n)
	return nil
}

type wechatAlbumArticle struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	MsgID    string `json:"msgid"`
	ItemIdx  string `json:"itemidx"`
}

func discoverWeChatAlbumLinks(cfg SourceConfig, maxPages int) (*ListDiscoverResult, error) {
	queue, pages, err := collectWeChatAlbumLinks(cfg, maxPages)
	if err != nil {
		return nil, err
	}
	sample := make([]string, 0, 10)
	for i, lk := range queue {
		if i >= 10 {
			break
		}
		sample = append(sample, lk.URL)
	}
	return &ListDiscoverResult{
		PageURLsScanned: pages,
		LinkCount:       len(queue),
		SampleURLs:      sample,
	}, nil
}
