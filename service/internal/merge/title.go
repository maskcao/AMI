package merge

import (
	"strings"
)

func NormalizeTitle(title string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(title), " "))
}

func TitlesMatch(a, b string) bool {
	return NormalizeTitle(a) == NormalizeTitle(b)
}

func FindKnowledgeID(incoming string, prior []struct {
	Title        string
	KnowledgeIDs []int64
}) int64 {
	for _, row := range prior {
		if !TitlesMatch(row.Title, incoming) {
			continue
		}
		if len(row.KnowledgeIDs) > 0 {
			return row.KnowledgeIDs[0]
		}
	}
	return 0
}
