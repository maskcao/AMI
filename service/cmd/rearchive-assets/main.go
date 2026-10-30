package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/ingest"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func main() {
	dataDir := flag.String("data", "./data", "AMI data directory")
	urlPattern := flag.String("url", "%ti.qianxin.com%", "SQL LIKE on intel_report.source_url")
	intelID := flag.Int64("intel-id", 0, "only process this intel id")
	dryRun := flag.Bool("dry-run", false, "report only, do not rewrite HTML")
	refetchSource := flag.Bool("refetch-source", false, "re-download source_url HTML before archiving images")
	invalidOnly := flag.Bool("invalid-only", false, "only rows whose raw-assets refs are missing or not real images")
	flag.Parse()

	db, err := store.Open(*dataDir)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	files := artifacts.New(*dataDir)
	q := `SELECT id, source_url FROM intel_report WHERE has_raw=1 AND source_url LIKE ?`
	args := []any{*urlPattern}
	if *intelID > 0 {
		q += ` AND id=?`
		args = append(args, *intelID)
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		log.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var fixed, skipped, failed int
	for rows.Next() {
		var id int64
		var sourceURL string
		if err := rows.Scan(&id, &sourceURL); err != nil {
			log.Fatalf("scan: %v", err)
		}
		raw, err := files.ReadRaw(id, 1)
		if err != nil {
			failed++
			continue
		}
		refs, valid, invalid := pipeline.AuditRawAssetRefs(files, id, raw)
		_, _, unresolved, remote := pipeline.AuditHTMLImages(raw)
		needs := unresolved > 0 || remote > 0 || invalid > 0
		if *invalidOnly && invalid == 0 && refs > 0 {
			skipped++
			continue
		}
		if !*invalidOnly && !*refetchSource && unresolved == 0 && remote == 0 && invalid == 0 {
			skipped++
			continue
		}
		if !needs && !*refetchSource {
			skipped++
			continue
		}
		if *dryRun {
			fmt.Printf("intel %d: img_refs=%d valid_assets=%d invalid_assets=%d unresolved_uploads=%d remote=%d refetch=%v\n",
				id, refs, valid, invalid, unresolved, remote, *refetchSource)
			fixed++
			continue
		}
		in := raw
		if *refetchSource {
			if strings.TrimSpace(sourceURL) == "" {
				log.Printf("intel %d: empty source_url", id)
				failed++
				continue
			}
			fresh, err := ingest.FetchArticleHTML(sourceURL, 45)
			if err != nil {
				log.Printf("intel %d refetch: %v", id, err)
				failed++
				continue
			}
			in = fresh
		}
		out, assetN, err := pipeline.ArchivePageAssets(files, id, sourceURL, in)
		if err != nil {
			log.Printf("intel %d archive: %v", id, err)
			failed++
			continue
		}
		if len(out) == 0 {
			out = in
		}
		if _, _, _, err := files.WriteRawHTML(id, 1, sourceURL, out); err != nil {
			log.Printf("intel %d write: %v", id, err)
			failed++
			continue
		}
		md := pipeline.HTMLToMarkdown(string(out))
		if md == "" {
			md = "(empty)"
		}
		if mdRel, err := files.WriteParsedMarkdown(id, "v1", md); err == nil {
			snip := pipeline.Snippet(md, 512)
			parseVer := fmt.Sprintf("html-v2+assets(%d)", assetN)
			_ = db.UpdateIntelAfterParse(id, snip, parseVer, mdRel)
		}
		_ = pipeline.SyncFiguresFromHTML(files, db, id, sourceURL, out)
		_, _, inv2 := pipeline.AuditRawAssetRefs(files, id, out)
		_, _, u, r := pipeline.AuditHTMLImages(out)
		fmt.Printf("intel %d: rearchived assets=%d invalid_raw_assets=%d remaining uploads=%d remote=%d\n", id, assetN, inv2, u, r)
		fixed++
	}
	log.Printf("done: fixed=%d skipped=%d failed=%d dry_run=%v refetch=%v invalid_only=%v pattern=%s",
		fixed, skipped, failed, *dryRun, *refetchSource, *invalidOnly, *urlPattern)
}
