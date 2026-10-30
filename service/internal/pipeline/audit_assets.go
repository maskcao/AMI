package pipeline

import (
	"fmt"
	"regexp"

	"github.com/ami/ami/service/internal/artifacts"
)

var rawAssetRefRe = regexp.MustCompile(`/api/v1/intel/(\d+)/raw-assets/([a-zA-Z0-9._-]+)`)

// AuditRawAssetRefs counts <img src> pointing at raw-assets and how many files are valid images on disk.
func AuditRawAssetRefs(files *artifacts.Store, intelID int64, html []byte) (refs, valid, invalid int) {
	if files == nil || intelID <= 0 {
		return 0, 0, 0
	}
	seen := map[string]struct{}{}
	for _, m := range imgSrcRe.FindAllStringSubmatch(string(html), -1) {
		if len(m) < 2 {
			continue
		}
		ref := m[1]
		sub := rawAssetRefRe.FindStringSubmatch(ref)
		if len(sub) != 3 {
			continue
		}
		if sub[1] != fmt.Sprintf("%d", intelID) {
			continue
		}
		name := sub[2]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		refs++
		body, err := files.ReadRawAsset(intelID, 1, name)
		if err != nil || !IsImageBytes(body) {
			invalid++
			continue
		}
		valid++
	}
	return refs, valid, invalid
}
