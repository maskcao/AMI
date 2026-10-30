package ingest

import (
	"fmt"
	"strconv"
	"strings"
)

// ExpandListPageURLs returns list page URLs to fetch (first page + path pagination).
func ExpandListPageURLs(cfg SourceConfig) []string {
	seed := cfg.ListURL
	if seed == "" {
		seed = cfg.EntryURL
	}
	if seed == "" {
		return nil
	}
	pag := cfg.Pagination
	if pag.Type == "" || pag.Type == "none" {
		return []string{seed}
	}
	if pag.Type == "scroll" {
		// Scroll pages share one URL; handled by render layer (not implemented).
		return []string{seed}
	}
	if pag.Type != "path" || pag.PathTemplate == "" {
		return []string{seed}
	}
	out := []string{seed}
	start := pag.StartPage
	if start <= 0 {
		start = 2
	}
	maxPages := pag.MaxPages
	if maxPages <= 0 {
		maxPages = 5
	}
	for i := 0; i < maxPages; i++ {
		page := start + i
		u := strings.ReplaceAll(pag.PathTemplate, "{page}", strconv.Itoa(page))
		u = strings.ReplaceAll(u, "{n}", strconv.Itoa(page))
		out = append(out, u)
	}
	return out
}

func ScrollPaginationError() error {
	return fmt.Errorf("pagination.type=scroll requires Playwright render (not enabled); for WeChat albums use pagination.type=wechat_album")
}
