package llm

import (
	"context"
	"os"
	"testing"

	"github.com/ami/ami/service/internal/store"
)

func TestEnrichIntegration(t *testing.T) {
	if os.Getenv("AMI_LLM_INTEGRATION") != "1" {
		t.Skip("set AMI_LLM_INTEGRATION=1 to run")
	}
	db, err := store.Open("../../data")
	if err != nil {
		t.Fatal(err)
	}
	out, err := EnrichWithDB(db, context.Background(), "ZHtrap botnet. CVE-2021-44228. MITRE T1110 brute force.")
	if err != nil {
		t.Fatal(err)
	}
	er := out.Result
	if er.Summary == "" {
		t.Fatal("empty summary")
	}
	t.Logf("provider=%s tag=%s summary=%q", out.Provider, out.VersionTag, er.Summary)
}
