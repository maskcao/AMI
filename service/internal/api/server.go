package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	DB    *store.DB
	Files *artifacts.Store
}

func New(db *store.DB, files *artifacts.Store) http.Handler {
	s := &Server{DB: db, Files: files}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(_ *http.Request, origin string) bool {
			if origin == "" {
				return true
			}
			if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" {
				return true
			}
			// LAN dev: Vite on 192.168.x.x:5173
			return strings.HasPrefix(origin, "http://192.168.") && strings.HasSuffix(origin, ":5173")
		},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/intel", s.listIntel)
		r.Post("/intel/import", s.importIntel)
		r.Get("/intel/{id}", s.getIntel)
		r.Get("/intel/{id}/original", s.getIntelOriginal)
		r.Get("/intel/{id}/parsed", s.getIntelParsed)
		r.Get("/intel/{id}/markdown", s.getIntelMarkdown)
		r.Get("/intel/{id}/raw-assets/{name}", s.getIntelRawAsset)
		r.Get("/intel/{id}/artifacts/{name}", s.getIntelArtifact)
		r.Get("/knowledge", s.listKnowledge)
		r.Get("/knowledge/{id}", s.getKnowledge)
		r.Patch("/knowledge/{id}", s.patchKnowledge)
		r.Post("/knowledge/{id}/publish", s.publishKnowledge)
		r.Post("/knowledge/merge", s.mergeKnowledge)
		r.Post("/intel/{id}/retry", s.retryIntel)
		r.Get("/ioc", s.searchIOC)
		r.Get("/analytics/bundle", s.analyticsBundle)
		r.Get("/analytics/actors/{slug}", s.actorProfile)
		r.Get("/settings/collect-sources", s.listCollectSources)
		r.Post("/settings/collect-sources", s.createCollectSource)
		r.Patch("/settings/collect-sources/{id}", s.patchCollectSource)
		r.Delete("/settings/collect-sources/{id}", s.deleteCollectSource)
		r.Post("/settings/collect-sources/{id}/run", s.runCollectSource)
		r.Get("/settings/model-providers", s.listModelProviders)
		r.Post("/settings/model-providers", s.createModelProvider)
		r.Patch("/settings/model-providers/{id}", s.patchModelProvider)
		r.Delete("/settings/model-providers/{id}", s.deleteModelProvider)
		r.Post("/settings/model-providers/{id}/test", s.testModelProvider)
		r.Get("/settings/parser-profiles", s.listParserProfiles)
		r.Post("/settings/parser-profiles", s.createParserProfile)
		r.Patch("/settings/parser-profiles/{id}", s.patchParserProfile)
		r.Delete("/settings/parser-profiles/{id}", s.deleteParserProfile)
		r.Post("/admin/purge-data", s.purgeAllData)
		r.Get("/datacenter/runtime-logs", s.listRuntimeLogs)
		r.Post("/datacenter/runtime-logs/clear", s.clearRuntimeLogs)
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request) int64 {
	id, _ := parseInt64(chi.URLParam(r, "id"))
	return id
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
