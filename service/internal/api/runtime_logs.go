package api

import (
	"net/http"
	"strconv"

	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/store"
)

func (s *Server) listRuntimeLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	rows, err := s.DB.ListRuntimeLogs(store.RuntimeLogListParams{
		Level:  q.Get("level"),
		Source: q.Get("source"),
		Query:  q.Get("q"),
		Limit:  limit,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sources, _ := s.DB.RuntimeLogSources()
	total, _ := s.DB.RuntimeLogCount()
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		m := map[string]any{
			"id":        row.ID,
			"createdAt": row.CreatedAt,
			"level":     row.Level,
			"source":    row.Source,
			"message":   row.Message,
		}
		if row.Detail != "" {
			m["detail"] = row.Detail
		}
		items = append(items, m)
	}
	if sources == nil {
		sources = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   items,
		"total":   total,
		"sources": sources,
	})
}

func (s *Server) clearRuntimeLogs(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.ClearRuntimeLogs(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	opslog.Info("system", "runtime logs cleared", "api")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
