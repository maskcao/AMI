package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) getIntelArtifact(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	name := chi.URLParam(r, "name")
	if id <= 0 || name == "" {
		writeErr(w, http.StatusBadRequest, "invalid id or name")
		return
	}
	name = filepath.Base(name)
	if strings.Contains(name, "..") {
		writeErr(w, http.StatusBadRequest, "invalid name")
		return
	}
	data, err := s.Files.ReadFigure(id, "v1", name)
	if err != nil {
		writeErr(w, http.StatusNotFound, "artifact not found")
		return
	}
	ct := "image/png"
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		ct = "image/jpeg"
	case ".gif":
		ct = "image/gif"
	case ".webp":
		ct = "image/webp"
	case ".svg":
		ct = "image/svg+xml"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}
