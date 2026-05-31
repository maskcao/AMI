package llm

import (
	"testing"

	"database/sql"
	"github.com/ami/ami/service/internal/store"
)

func TestProviderSortKeyPrefersEndpointID(t *testing.T) {
	auto := store.ModelProviderRow{
		Enabled: true, Kind: "llm",
		ModelID:    sql.NullString{String: "auto", Valid: true},
		APIKey:     sql.NullString{String: "k", Valid: true},
		DefaultFor: sql.NullString{String: "parse, enrich", Valid: true},
	}
	ep := store.ModelProviderRow{
		Enabled: true, Kind: "llm",
		ModelID:    sql.NullString{String: "ep-20260101123456-abcd", Valid: true},
		APIKey:     sql.NullString{String: "k", Valid: true},
		DefaultFor: sql.NullString{String: "parse, enrich", Valid: true},
	}
	if providerSortKey(auto) >= providerSortKey(ep) {
		t.Fatalf("ep should rank above auto: auto=%d ep=%d", providerSortKey(auto), providerSortKey(ep))
	}
}
