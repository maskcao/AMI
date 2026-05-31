package store

import (
	"encoding/json"
	"errors"
)

type KnowledgePatch struct {
	Title             *string
	Summary           *string
	TargetIndustries  []string
	Status            *string
	NeedsReview       *bool
}

func (d *DB) PatchKnowledge(id int64, patch KnowledgePatch) error {
	row, err := d.GetKnowledge(id)
	if err != nil {
		return err
	}
	if row == nil {
		return errors.New("knowledge not found")
	}
	title := row.Title
	if patch.Title != nil {
		title = *patch.Title
	}
	summary := row.Summary.String
	if patch.Summary != nil {
		summary = *patch.Summary
	}
	indJSON := row.TargetIndustriesJSON
	if patch.TargetIndustries != nil {
		b, _ := json.Marshal(patch.TargetIndustries)
		indJSON = string(b)
	}
	status := row.Status
	if patch.Status != nil {
		status = *patch.Status
	}
	nr := 0
	if row.NeedsReview {
		nr = 1
	}
	if patch.NeedsReview != nil {
		if *patch.NeedsReview {
			nr = 1
		} else {
			nr = 0
		}
	}
	score := ComputeQualityScore(title, summary, indJSON, row.PayloadJSON, row.IntelCount)
	_, err = d.Exec(`
		UPDATE attack_knowledge SET title=?, summary=?, target_industries_json=?, status=?, needs_review=?, quality_score=?, updated_at=? WHERE id=?`,
		title, summary, indJSON, status, nr, score, nowRFC(), id)
	return err
}

func (d *DB) PublishKnowledge(id int64) error {
	status := "published"
	nr := false
	return d.PatchKnowledge(id, KnowledgePatch{Status: &status, NeedsReview: &nr})
}

func ComputeQualityScore(title, summary, industriesJSON, payloadJSON string, intelCount int) int {
	score := 0
	if len(title) > 5 {
		score += 15
	}
	if summary != "" {
		score += 25
	}
	var ind []string
	_ = json.Unmarshal([]byte(industriesJSON), &ind)
	if len(ind) > 0 {
		score += 20
	}
	pl, _ := ParsePayload(payloadJSON)
	if len(pl.Techniques) > 0 {
		score += 25
	}
	if intelCount > 0 {
		score += 15
	}
	if score > 100 {
		return 100
	}
	return score
}

func (d *DB) MergeKnowledge(sourceID, targetID int64) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT intel_id, role FROM knowledge_intel_link WHERE knowledge_id=?`, sourceID)
	if err != nil {
		return err
	}
	type link struct {
		intelID int64
		role    string
	}
	var links []link
	for rows.Next() {
		var l link
		if err := rows.Scan(&l.intelID, &l.role); err != nil {
			rows.Close()
			return err
		}
		links = append(links, l)
	}
	rows.Close()

	for _, l := range links {
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO knowledge_intel_link (knowledge_id, intel_id, role, linked_at)
			VALUES (?, ?, ?, ?)`, targetID, l.intelID, l.role, nowRFC())
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`DELETE FROM knowledge_intel_link WHERE knowledge_id=?`, sourceID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM attack_knowledge WHERE id=?`, sourceID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE attack_knowledge SET intel_count = (SELECT COUNT(*) FROM knowledge_intel_link WHERE knowledge_id=?), updated_at=? WHERE id=?`,
		targetID, nowRFC(), targetID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
