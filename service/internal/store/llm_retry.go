package store

// LLMFailedIntel returns intel rows that finished the pipeline but were
// enriched by the rules fallback instead of an LLM, oldest-refresh first.
func (d *DB) LLMFailedIntel(limit int) (ids []int64, remaining int, err error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := d.Query(`
		SELECT id FROM intel_report
		WHERE ingest_status = 'succeeded' AND (enrich_version = 'rules-v1' OR enrich_version IS NULL)
		ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	err = d.QueryRow(`
		SELECT COUNT(*) FROM intel_report
		WHERE ingest_status = 'succeeded' AND (enrich_version = 'rules-v1' OR enrich_version IS NULL)`).
		Scan(&remaining)
	return ids, remaining, err
}

// RequeueLLMFailed resets the given intel rows back to the extract stage so the
// worker retries LLM enrichment (knowledge is updated in place, not recreated).
func (d *DB) RequeueLLMFailed(ids []int64) error {
	for _, id := range ids {
		row, err := d.GetIntel(id)
		if err != nil || row == nil {
			continue
		}
		if err := d.ClearPendingPipelineJobs(id); err != nil {
			return err
		}
		if err := d.SetIngestStatus(id, "queued", row.KnowledgeAction, nil, "", row.EnrichVersion.String, nil); err != nil {
			return err
		}
		if err := d.EnqueueJob(id, "extract"); err != nil {
			return err
		}
	}
	return nil
}
