package merge

import (
	"strings"
	"unicode/utf8"
)

func NormalizeTitle(title string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(title), " "))
}

func TitlesMatch(a, b string) bool {
	return NormalizeTitle(a) == NormalizeTitle(b)
}

var nonMergeableTitles = map[string]struct{}{
	"奇安信威胁情报中心":                      {},
	"qianxin threat intelligence center": {},
}

// ShouldMergeByTitle gates knowledge merge: generic site titles must not collapse distinct articles.
func ShouldMergeByTitle(title string) bool {
	t := NormalizeTitle(title)
	if t == "" {
		return false
	}
	n := utf8.RuneCountInString(t)
	if n < 4 || n > 180 {
		return false
	}
	if _, ok := nonMergeableTitles[t]; ok {
		return false
	}
	if _, ok := nonMergeableTitles[strings.ToLower(t)]; ok {
		return false
	}
	return true
}

func FindKnowledgeID(incoming string, prior []struct {
	Title        string
	KnowledgeIDs []int64
}) int64 {
	if !ShouldMergeByTitle(incoming) {
		return 0
	}
	for _, row := range prior {
		if !TitlesMatch(row.Title, incoming) {
			continue
		}
		if !ShouldMergeByTitle(row.Title) {
			continue
		}
		if len(row.KnowledgeIDs) > 0 {
			return row.KnowledgeIDs[0]
		}
	}
	return 0
}
