package analytics

import (
	"sync"

	"github.com/ami/ami/service/internal/store"
)

// bundleCache wraps BuildBundle with an in-process cache invalidated by a
// lightweight watermark (row count + max updated_at) over the two source
// tables, so repeated /analytics/bundle requests skip the full aggregation.
type bundleWatermark struct {
	knowledgeCount int64
	knowledgeMax   string
	intelCount     int64
	intelMax       string
}

var (
	bundleMu    sync.Mutex
	bundleCache *Bundle
	bundleWM    bundleWatermark
)

func currentWatermark(db *store.DB) (bundleWatermark, error) {
	var wm bundleWatermark
	if err := db.QueryRow(`SELECT COUNT(*), COALESCE(MAX(updated_at), '') FROM attack_knowledge`).
		Scan(&wm.knowledgeCount, &wm.knowledgeMax); err != nil {
		return wm, err
	}
	if err := db.QueryRow(`SELECT COUNT(*), COALESCE(MAX(updated_at), '') FROM intel_report`).
		Scan(&wm.intelCount, &wm.intelMax); err != nil {
		return wm, err
	}
	return wm, nil
}

func BuildBundleCached(db *store.DB) (*Bundle, error) {
	wm, wmErr := currentWatermark(db)
	if wmErr == nil {
		bundleMu.Lock()
		hit := bundleCache != nil && wm == bundleWM
		cached := bundleCache
		bundleMu.Unlock()
		if hit {
			return cached, nil
		}
	}
	b, err := BuildBundle(db)
	if err != nil {
		return nil, err
	}
	if wmErr == nil {
		bundleMu.Lock()
		bundleCache = b
		bundleWM = wm
		bundleMu.Unlock()
	}
	return b, nil
}
