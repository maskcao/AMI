package api

import (
	"encoding/json"
	"net/http"
)

var collectTypeLabels = map[string]string{
	"direct_url":      "单 URL 直采",
	"paginated_list":  "列表监测（路径分页）",
	"threat_portal":   "威胁门户（列表抽链）",
	"scheduled_url":   "定时 URL 监测",
	"wechat_feed":     "微信文章采集",
	"rss":             "RSS / Atom",
	"manual_upload":   "手工上传",
}

func (s *Server) listCollectSources(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.ListCollectSources()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		full, _ := s.DB.GetCollectSource(r.ID)
		if full != nil {
			items = append(items, collectSourceJSON(*full))
			continue
		}
		m := map[string]any{
			"id": r.ID, "name": r.Name, "type": r.Type,
			"typeLabel": collectTypeLabels[r.Type],
			"enabled": r.Enabled, "sourceLabel": r.SourceLabel,
			"config": json.RawMessage("{}"),
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
		items = append(items, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) listModelProviders(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.ListModelProviders()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		items = append(items, modelProviderJSON(r))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) listParserProfiles(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.DB.ListParserProfiles()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		m := map[string]any{"id": r.ID, "name": r.Name, "usedBySources": r.UsedBySources}
		if r.Description.Valid {
			m["description"] = r.Description.String
		}
		items = append(items, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
