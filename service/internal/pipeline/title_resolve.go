package pipeline

import (
	"encoding/json"

	"github.com/ami/ami/service/internal/store"
)

type collectTitleConfig struct {
	TitleExtract   []TitleExtractRule `json:"title_extract"`
	TitleBlocklist []string           `json:"title_blocklist"`
}

// TitleOptsForIntel loads per-source title_extract from collect_source and applies URL defaults.
func TitleOptsForIntel(db *store.DB, collectSourceName, sourceURL string) *TitleExtractOpts {
	var fromCollect *TitleExtractOpts
	if collectSourceName != "" {
		src, err := db.GetCollectSourceByName(collectSourceName)
		if err == nil && src != nil && src.ConfigJSON != "" {
			var patch collectTitleConfig
			if json.Unmarshal([]byte(src.ConfigJSON), &patch) == nil {
				if len(patch.TitleExtract) > 0 || len(patch.TitleBlocklist) > 0 {
					fromCollect = &TitleExtractOpts{
						Rules:     patch.TitleExtract,
						Blocklist: patch.TitleBlocklist,
					}
				}
			}
		}
	}
	return TitleOptsForURL(sourceURL, fromCollect)
}
