package api

import (
	"net/http"
)

func (s *Server) getIntelParsed(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetIntel(id)
	if err != nil || row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	bodyPath, _ := s.DB.ParsedLLMInputPath(id)
	if bodyPath != "" {
		b, err := s.Files.ReadParsedText(s.DB.DataDir(), bodyPath)
		if err == nil && len(b) > 0 {
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			_, _ = w.Write(b)
			return
		}
	}
	if row.BodyPreview.Valid && row.BodyPreview.String != "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(row.BodyPreview.String))
		return
	}
	writeErr(w, http.StatusNotFound, "no parsed body")
}
