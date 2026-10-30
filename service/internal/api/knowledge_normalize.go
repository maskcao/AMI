package api

import (
	"strings"

	"github.com/ami/ami/service/internal/malwareclass"
)

func normalizeMalwarePayload(items []map[string]any) []map[string]any {
	if len(items) == 0 {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(items))
	for _, m := range items {
		name, _ := m["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		cls := malwareclass.ClassFromMap(m)
		row := map[string]any{"name": name}
		if cls != "" {
			row["class"] = cls
		}
		out = append(out, row)
	}
	return out
}
