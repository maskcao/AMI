package corpus

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ami/ami/service/internal/store"
)

func ExportOnPublish(db *store.DB, knowledgeID int64) (string, error) {
	row, err := db.GetKnowledge(knowledgeID)
	if err != nil || row == nil || row.Status != "published" {
		return "", fmt.Errorf("knowledge not published")
	}
	pl, _ := store.ParsePayload(row.PayloadJSON)
	links, _ := db.LinkedIntelForKnowledge(knowledgeID)
	var industries []string
	_ = json.Unmarshal([]byte(row.TargetIndustriesJSON), &industries)

	evidence := make([]map[string]any, 0, len(links))
	for _, l := range links {
		item := map[string]any{"intelId": l.ID, "title": l.Title, "role": l.Role}
		if l.DisclosedAt.Valid {
			item["disclosedAt"] = l.DisclosedAt.String
		}
		evidence = append(evidence, item)
	}

	record := map[string]any{
		"knowledgeId":       knowledgeID,
		"title":             row.Title,
		"summary":           row.Summary.String,
		"targetIndustries":  industries,
		"techniques":        pl.Techniques,
		"procedures":        pl.Procedures,
		"evidence":          evidence,
		"schemaVersion":     "1",
		"exportedAt":        time.Now().UTC().Format(time.RFC3339),
	}
	line, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	dir := filepath.Join(db.DataDir(), "store", "corpus", "knowledge", fmt.Sprintf("%d", knowledgeID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ts := time.Now().UTC().Format("20060102150405")
	rel := filepath.ToSlash(filepath.Join("store", "corpus", "knowledge", fmt.Sprintf("%d", knowledgeID), "snapshot-"+ts+".jsonl"))
	abs := filepath.Join(db.DataDir(), rel)
	if err := os.WriteFile(abs, append(line, '\n'), 0o644); err != nil {
		return "", err
	}
	_, err = db.Exec(`
		INSERT INTO knowledge_corpus_export (knowledge_id, path, format, created_at)
		VALUES (?, ?, 'jsonl', ?)`, knowledgeID, rel, time.Now().UTC().Format(time.RFC3339))
	return rel, err
}
