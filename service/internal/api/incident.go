package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/store"
)

const incidentMaxUpload = 50 << 20 // 50MB

func readAllLimited(f io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("file too large")
	}
	return data, nil
}

func (s *Server) uploadIncidentReport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, incidentMaxUpload)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "上传解析失败，文件可能超过 50MB 限制")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "缺少 file 字段")
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".md" && ext != ".markdown" && ext != ".docx" {
		writeErr(w, http.StatusBadRequest, "仅支持 .md / .docx 文件")
		return
	}
	data, err := readAllLimited(file, incidentMaxUpload)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "文件读取失败或超过大小限制")
		return
	}
	if len(data) == 0 {
		writeErr(w, http.StatusBadRequest, "文件内容为空")
		return
	}
	if ext == ".docx" && !strings.HasPrefix(string(data[:4]), "PK") {
		writeErr(w, http.StatusBadRequest, "不是有效的 .docx 文件")
		return
	}

	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])

	created, err := s.DB.CreateIncidentReport(
		strings.TrimSuffix(filename, filepath.Ext(filename)),
		filename, ext, header.Header.Get("Content-Type"), int64(len(data)), sha, "", r.FormValue("reportTime"),
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "保存记录失败")
		return
	}
	relPath, err := s.Files.WriteIncidentFile(created.ID, filename, data)
	if err != nil {
		_ = s.DB.DeleteIncidentReport(created.ID)
		writeErr(w, http.StatusInternalServerError, "保存文件失败")
		return
	}
	if _, err := s.DB.Exec(`UPDATE incident_report SET storage_path = ? WHERE id = ?`, relPath, created.ID); err != nil {
		writeErr(w, http.StatusInternalServerError, "更新存储路径失败")
		return
	}
	created.StoragePath = relPath
	writeJSON(w, http.StatusOK, created)
}

func (s *Server) listIncidentReports(w http.ResponseWriter, _ *http.Request) {
	items, err := s.DB.ListIncidentReports()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "查询失败")
		return
	}
	if items == nil {
		items = []*store.IncidentReport{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *Server) getIncidentReport(w http.ResponseWriter, r *http.Request) {
	it, ok := s.loadIncident(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) deleteIncidentReport(w http.ResponseWriter, r *http.Request) {
	it, ok := s.loadIncident(w, r)
	if !ok {
		return
	}
	if err := s.DB.DeleteIncidentReport(it.ID); err != nil {
		writeErr(w, http.StatusInternalServerError, "删除失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) patchIncidentReport(w http.ResponseWriter, r *http.Request) {
	it, ok := s.loadIncident(w, r)
	if !ok {
		return
	}
	var body struct {
		ReportTime string `json:"reportTime"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败")
		return
	}
	body.ReportTime = strings.TrimSpace(body.ReportTime)
	if body.ReportTime != "" {
		if t, err := time.Parse("2006-01-02", body.ReportTime); err != nil || t.Year() < 1990 || t.Year() > 2100 {
			writeErr(w, http.StatusBadRequest, "报告时间格式应为 YYYY-MM-DD")
			return
		}
	}
	if err := s.DB.UpdateIncidentReportTime(it.ID, body.ReportTime); err != nil {
		writeErr(w, http.StatusInternalServerError, "更新失败")
		return
	}
	updated, err := s.DB.GetIncidentReport(it.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "查询失败")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// serveIncidentFile streams the stored file; inline=true renders in browser (preview), otherwise attachment download.
func (s *Server) serveIncidentFile(w http.ResponseWriter, r *http.Request, inline bool) {
	it, ok := s.loadIncident(w, r)
	if !ok {
		return
	}
	data, err := s.Files.ReadIncidentFile(it.StoragePath)
	if err != nil {
		writeErr(w, http.StatusNotFound, "文件不存在")
		return
	}
	ct := it.Mime
	switch it.Ext {
	case ".md", ".markdown":
		ct = "text/markdown; charset=utf-8"
	case ".docx":
		ct = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	w.Header().Set("Content-Type", ct)
	disposition := "attachment"
	if inline {
		disposition = "inline"
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`%s; filename*=UTF-8''%s`, disposition, url.PathEscape(it.Filename)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

func (s *Server) previewIncidentReport(w http.ResponseWriter, r *http.Request) {
	s.serveIncidentFile(w, r, true)
}

func (s *Server) downloadIncidentReport(w http.ResponseWriter, r *http.Request) {
	s.serveIncidentFile(w, r, false)
}

func (s *Server) loadIncident(w http.ResponseWriter, r *http.Request) (*store.IncidentReport, bool) {
	it, err := s.DB.GetIncidentReport(pathID(r))
	if err != nil {
		writeErr(w, http.StatusNotFound, "报告不存在")
		return nil, false
	}
	return it, true
}
