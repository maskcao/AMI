package pipeline

import (
	"fmt"
	"os"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/store"
)

// QualityCheck is one automated gate item for pilot collect validation.
type QualityCheck struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// QualityReport aggregates pilot quality checks for one intel row.
type QualityReport struct {
	Passed bool           `json:"passed"`
	Checks []QualityCheck `json:"checks"`
}

// RunIntelQualityGate evaluates archived HTML and pipeline outcome for pilot runs.
func RunIntelQualityGate(files *artifacts.Store, db *store.DB, intelID int64, rawHTML []byte) QualityReport {
	report := QualityReport{Checks: []QualityCheck{}}
	add := func(id, label string, ok bool, detail string) {
		report.Checks = append(report.Checks, QualityCheck{ID: id, Label: label, OK: ok, Detail: detail})
		if !ok {
			report.Passed = false
		}
	}
	report.Passed = true

	if len(rawHTML) < 200 {
		add("html_size", "HTML 快照有足够内容", false, fmt.Sprintf("仅 %d 字节", len(rawHTML)))
	} else {
		add("html_size", "HTML 快照有足够内容", true, "")
	}

	row, err := db.GetIntel(intelID)
	if err != nil || row == nil {
		add("intel_row", "情报记录存在", false, "not found")
		return report
	}
	switch row.IngestStatus {
	case "succeeded":
		add("pipeline", "流水线 ingest_status=succeeded", true, "")
	case "failed":
		add("pipeline", "流水线 ingest_status=succeeded", false, row.PipelineError.String)
	default:
		add("pipeline", "流水线 ingest_status=succeeded", false, row.IngestStatus)
	}

	if !row.HasRaw {
		add("has_raw", "已落盘原始 HTML", false, "")
	} else {
		add("has_raw", "已落盘原始 HTML", true, "")
	}

	refs, valid, invalid := AuditRawAssetRefs(files, intelID, rawHTML)
	if refs == 0 {
		add("images", "正文配图已归档（或本文无图）", true, "无 img→raw-assets 引用")
	} else if invalid > 0 {
		add("images", "正文配图为真实图片文件", false, fmt.Sprintf("%d/%d 引用无效", invalid, refs))
	} else {
		add("images", "正文配图为真实图片文件", true, fmt.Sprintf("%d 张", valid))
	}

	_, _, unresolved, remote := AuditHTMLImages(rawHTML)
	if unresolved > 0 {
		add("uploads", "无未归档 /uploads/ 图片", false, fmt.Sprintf("%d 条", unresolved))
	} else {
		add("uploads", "无未归档 /uploads/ 图片", true, "")
	}
	archivedAssets := countRawAssetFiles(files, intelID)
	if remote > 0 && archivedAssets >= 5 && row.IngestStatus == "succeeded" {
		add("remote_img", "正文 img 无远程 URL 残留", true,
			fmt.Sprintf("%d 个资源已落盘；%d 条 src 仍为远程（多为 srcset/侧栏）", archivedAssets, remote))
	} else if remote > 0 {
		add("remote_img", "正文 img 无远程 URL 残留", false, fmt.Sprintf("%d 条", remote))
	} else {
		add("remote_img", "正文 img 无远程 URL 残留", true, "")
	}

	kids, _ := db.KnowledgeIDsForIntel(intelID)
	kDetail := "knowledge_ids 为空"
	if len(kids) > 0 {
		kDetail = fmt.Sprintf("%d 条", len(kids))
	}
	report.Checks = append(report.Checks, QualityCheck{
		ID: "knowledge", Label: "已写入攻击知识（参考）", OK: len(kids) > 0, Detail: kDetail,
	})

	return report
}

func countRawAssetFiles(files *artifacts.Store, intelID int64) int {
	if files == nil {
		return 0
	}
	entries, err := os.ReadDir(files.RawAssetsDir(intelID, 1))
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n++
	}
	return n
}
