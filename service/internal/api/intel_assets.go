package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) getIntelRawAsset(w http.ResponseWriter, r *http.Request) {
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
	data, err := s.Files.ReadRawAsset(id, 1, name)
	if err != nil {
		writeErr(w, http.StatusNotFound, "asset not found")
		return
	}
	ct := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(name)) {
	case ".css":
		ct = "text/css; charset=utf-8"
	case ".js":
		ct = "application/javascript; charset=utf-8"
	case ".jpg", ".jpeg":
		ct = "image/jpeg"
	case ".png":
		ct = "image/png"
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
