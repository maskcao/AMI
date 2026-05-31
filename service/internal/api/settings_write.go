package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ami/ami/service/internal/ingest"
	"github.com/ami/ami/service/internal/llm"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) createCollectSource(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name          string          `json:"name"`
		Type          string          `json:"type"`
		Enabled       *bool           `json:"enabled"`
		ScheduleCron  string          `json:"scheduleCron"`
		ScheduleLabel string          `json:"scheduleLabel"`
		SourceLabel   string          `json:"sourceLabel"`
		Config        json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.Type) == "" {
		writeErr(w, http.StatusBadRequest, "name and type required")
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	cfg := "{}"
	if len(body.Config) > 0 {
		cfg = string(body.Config)
	}
	id, err := s.DB.CreateCollectSource(store.CollectSourceCreate{
		Name: body.Name, Type: body.Type, Enabled: enabled,
		ScheduleCron: body.ScheduleCron, ScheduleLabel: body.ScheduleLabel,
		SourceLabel: body.SourceLabel, ConfigJSON: cfg,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	src, _ := s.DB.GetCollectSource(id)
	if src == nil {
		writeErr(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJSON(w, http.StatusCreated, collectSourceJSON(*src))
}

func (s *Server) patchCollectSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name          *string         `json:"name"`
		Enabled       *bool           `json:"enabled"`
		ScheduleCron  *string         `json:"scheduleCron"`
		ScheduleLabel *string         `json:"scheduleLabel"`
		SourceLabel   *string         `json:"sourceLabel"`
		Config        json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	patch := store.CollectSourcePatch{
		Name: body.Name, Enabled: body.Enabled,
		ScheduleCron: body.ScheduleCron, ScheduleLabel: body.ScheduleLabel,
		SourceLabel: body.SourceLabel,
	}
	if len(body.Config) > 0 {
		cfg := string(body.Config)
		patch.ConfigJSON = &cfg
	}
	if err := s.DB.UpdateCollectSource(id, patch); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	src, _ := s.DB.GetCollectSource(id)
	if src == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, collectSourceJSON(*src))
}

func (s *Server) createModelProvider(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		Kind       string `json:"kind"`
		ModelID    string `json:"modelId"`
		BaseURL    string `json:"baseUrl"`
		APIKey     string `json:"apiKey"`
		Enabled    *bool  `json:"enabled"`
		DefaultFor string `json:"defaultFor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeErr(w, http.StatusBadRequest, "name required")
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	id, err := s.DB.CreateModelProvider(store.ModelProviderCreate{
		Name: body.Name, Kind: body.Kind, ModelID: body.ModelID, BaseURL: body.BaseURL,
		APIKey: body.APIKey, DefaultFor: body.DefaultFor, Enabled: enabled,
	})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	row, _ := s.DB.GetModelProvider(id)
	if row == nil {
		writeErr(w, http.StatusInternalServerError, "create failed")
		return
	}
	writeJSON(w, http.StatusCreated, modelProviderJSON(*row))
}

func (s *Server) createParserProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	id, err := s.DB.CreateParserProfile(store.ParserProfileCreate{Name: body.Name, Description: body.Description})
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, _ := s.DB.ListParserProfiles()
	for _, row := range rows {
		if row.ID == id {
			m := map[string]any{"id": row.ID, "name": row.Name, "usedBySources": row.UsedBySources}
			if row.Description.Valid {
				m["description"] = row.Description.String
			}
			writeJSON(w, http.StatusCreated, m)
			return
		}
	}
	writeErr(w, http.StatusInternalServerError, "create failed")
}

func (s *Server) testModelProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	row, err := s.DB.GetModelProvider(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if err := llm.TestProvider(r.Context(), row); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "providerId": id, "name": row.Name})
}

func (s *Server) patchModelProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name       *string `json:"name"`
		ModelID    *string `json:"modelId"`
		BaseURL    *string `json:"baseUrl"`
		APIKey     *string `json:"apiKey"`
		Enabled    *bool   `json:"enabled"`
		DefaultFor *string `json:"defaultFor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.DB.UpdateModelProvider(id, store.ModelProviderPatch{
		Name: body.Name, ModelID: body.ModelID, BaseURL: body.BaseURL, APIKey: body.APIKey,
		Enabled: body.Enabled, DefaultFor: body.DefaultFor,
	}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	row, _ := s.DB.GetModelProvider(id)
	if row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, modelProviderJSON(*row))
}

func (s *Server) runCollectSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := s.DB.GetCollectSource(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if src == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	opslog.Info("api", "manual collect run", "sourceId="+id)
	res, err := ingest.RunCollectSource(s.DB, s.Files, id)
	out := map[string]any{"sourceId": id}
	if res != nil {
		out["imported"] = res.Imported
		out["skipped"] = res.Skipped
		out["failed"] = res.Failed
	}
	if err != nil {
		out["error"] = err.Error()
		writeJSON(w, http.StatusOK, out)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) deleteCollectSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.DB.DeleteCollectSource(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteModelProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.DB.DeleteModelProvider(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) patchParserProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.DB.UpdateParserProfile(id, store.ParserProfilePatch{Name: body.Name, Description: body.Description}); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, _ := s.DB.ListParserProfiles()
	for _, row := range rows {
		if row.ID == id {
			m := map[string]any{"id": row.ID, "name": row.Name, "usedBySources": row.UsedBySources}
			if row.Description.Valid {
				m["description"] = row.Description.String
			}
			writeJSON(w, http.StatusOK, m)
			return
		}
	}
	writeErr(w, http.StatusNotFound, "not found")
}

func (s *Server) deleteParserProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.DB.DeleteParserProfile(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func collectSourceJSON(r store.CollectSourceFull) map[string]any {
	m := map[string]any{
		"id": r.ID, "name": r.Name, "type": r.Type,
		"typeLabel": collectTypeLabels[r.Type],
		"enabled": r.Enabled, "sourceLabel": r.SourceLabel,
		"config": json.RawMessage(r.ConfigJSON),
	}
	if r.ScheduleCron.Valid {
		m["scheduleCron"] = r.ScheduleCron.String
	}
	if r.ScheduleLabel.Valid {
		m["scheduleLabel"] = r.ScheduleLabel.String
	}
	if r.LastRunAt.Valid {
		m["lastRunAt"] = r.LastRunAt.String
	}
	if r.LastStatus.Valid {
		m["lastStatus"] = r.LastStatus.String
	}
	return m
}

func modelProviderJSON(r store.ModelProviderRow) map[string]any {
	m := map[string]any{
		"id": r.ID, "name": r.Name, "kind": r.Kind, "enabled": r.Enabled,
		"apiKeyConfigured": r.APIKey.Valid && r.APIKey.String != "",
	}
	if r.ModelID.Valid {
		m["modelId"] = r.ModelID.String
	}
	if r.BaseURL.Valid {
		m["baseUrl"] = r.BaseURL.String
	}
	if r.DefaultFor.Valid {
		m["defaultFor"] = r.DefaultFor.String
	}
	return m
}
