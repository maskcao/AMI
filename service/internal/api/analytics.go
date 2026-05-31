package api

import (
	"net/http"

	"github.com/ami/ami/service/internal/analytics"
	"github.com/go-chi/chi/v5"
)

func (s *Server) analyticsBundle(w http.ResponseWriter, r *http.Request) {
	b, err := analytics.BuildBundle(s.DB)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) actorProfile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	b, err := analytics.BuildBundle(s.DB)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	profiles, _ := b.ActorProfiles.(map[string]any)
	if p, ok := profiles[slug]; ok {
		writeJSON(w, http.StatusOK, p)
		return
	}
	writeErr(w, http.StatusNotFound, "actor not found")
}
