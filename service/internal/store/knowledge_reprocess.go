package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// DeleteKnowledgeLinkedToIntelURL removes attack_knowledge rows linked to intel matching sourceURLPattern (SQL LIKE).
func (d *DB) DeleteKnowledgeLinkedToIntelURL(sourceURLPattern string) (knowledgeDeleted, intelCount int, err error) {
	intelIDs, err := d.intelIDsBySourceURLPattern(sourceURLPattern)
	if err != nil {
		return 0, 0, err
	}
	if len(intelIDs) == 0 {
		return 0, 0, nil
	}
	knowledgeIDs, err := d.knowledgeIDsForIntelIDs(intelIDs)
	if err != nil {
		return 0, 0, err
	}

	tx, err := d.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	if err := execDeleteIn(tx, "knowledge_intel_link", "intel_id", intelIDs); err != nil {
		return 0, 0, err
	}
	if len(knowledgeIDs) > 0 {
		if err := execDeleteIn(tx, "attack_knowledge", "id", knowledgeIDs); err != nil {
			return 0, 0, err
		}
	}
	if err := execDeleteIn(tx, "pipeline_jobs", "intel_id", intelIDs); err != nil {
		return 0, 0, err
	}
	if err := resetIntelForReprocess(tx, intelIDs); err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return len(knowledgeIDs), len(intelIDs), nil
}

func resetIntelForReprocess(tx *sql.Tx, intelIDs []int64) error {
	q := fmt.Sprintf(`UPDATE intel_report SET ingest_status='queued', knowledge_action='pending',
		extract_confidence=NULL, pipeline_error=NULL, enrich_version=NULL, ai_extract_json=NULL, updated_at=?
		WHERE id IN (%s)`, inClause(len(intelIDs)))
	args := make([]any, 0, len(intelIDs)+1)
	args = append(args, nowRFC())
	for _, id := range intelIDs {
		args = append(args, id)
	}
	_, err := tx.Exec(q, args...)
	return err
}

// EnqueueIntelReprocess queues parse jobs for intel matching source URL pattern.
func (d *DB) EnqueueIntelReprocess(sourceURLPattern string) (int, error) {
	ids, err := d.intelIDsBySourceURLPattern(sourceURLPattern)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		if err := d.EnqueueJob(id, "parse"); err != nil {
			return 0, fmt.Errorf("intel %d: %w", id, err)
		}
	}
	return len(ids), nil
}

func (d *DB) intelIDsBySourceURLPattern(pattern string) ([]int64, error) {
	rows, err := d.Query(`SELECT id FROM intel_report WHERE source_url LIKE ? ORDER BY id`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (d *DB) knowledgeIDsForIntelIDs(intelIDs []int64) ([]int64, error) {
	if len(intelIDs) == 0 {
		return nil, nil
	}
	q := fmt.Sprintf(`SELECT DISTINCT knowledge_id FROM knowledge_intel_link WHERE intel_id IN (%s)`, inClause(len(intelIDs)))
	rows, err := d.Query(q, int64Args(intelIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func execDeleteIn(tx *sql.Tx, table, column string, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	q := fmt.Sprintf(`DELETE FROM %s WHERE %s IN (%s)`, table, column, inClause(len(ids)))
	_, err := tx.Exec(q, int64Args(ids)...)
	return err
}

func inClause(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ",")
}

func int64Args(ids []int64) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

// DeleteKnowledgeByID removes one knowledge row and its links, then re-queues extract for linked intel.
func (d *DB) DeleteKnowledgeByIDAndReextract(knowledgeID int64) (int, error) {
	rows, err := d.Query(`SELECT intel_id FROM knowledge_intel_link WHERE knowledge_id=?`, knowledgeID)
	if err != nil {
		return 0, err
	}
	var intelIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		intelIDs = append(intelIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	tx, err := d.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`DELETE FROM knowledge_intel_link WHERE knowledge_id=?`, knowledgeID)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`DELETE FROM attack_knowledge WHERE id=?`, knowledgeID)
	if err != nil {
		return 0, err
	}
	if len(intelIDs) > 0 {
		if err := execDeleteIn(tx, "pipeline_jobs", "intel_id", intelIDs); err != nil {
			return 0, err
		}
		if err := resetIntelForReprocess(tx, intelIDs); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	for _, id := range intelIDs {
		if err := d.EnqueueJob(id, "parse"); err != nil {
			return 0, err
		}
	}
	return len(intelIDs), nil
}

// RequeueIntelStages clears jobs and enqueues parse for the given intel ids.
func (d *DB) RequeueIntelStages(intelIDs []int64) error {
	if len(intelIDs) == 0 {
		return nil
	}
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := execDeleteIn(tx, "pipeline_jobs", "intel_id", intelIDs); err != nil {
		return err
	}
	if err := resetIntelForReprocess(tx, intelIDs); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, id := range intelIDs {
		if err := d.EnqueueJob(id, "parse"); err != nil {
			return err
		}
	}
	return nil
}
