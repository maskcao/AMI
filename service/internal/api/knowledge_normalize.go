package api

import "strings"

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
		cls, _ := m["class"].(string)
		if strings.TrimSpace(cls) == "" {
			if c2, ok := m["classification"].(string); ok {
				cls = strings.TrimSpace(c2)
			}
		}
		row := map[string]any{"name": name}
		if cls != "" {
			row["class"] = cls
		}
		out = append(out, row)
	}
	return out
}
