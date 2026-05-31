package llm

import "testing"

func TestValidateEnrichRejectsNoisySummary(t *testing.T) {
	er := &EnrichResult{
		Summary: `{ "@context": "https://schema.org"`,
	}
	q := ValidateEnrich(er)
	if q.OK {
		t.Fatal("expected fail")
	}
}

func TestValidateEnrichAcceptsGoodSummary(t *testing.T) {
	er := &EnrichResult{
		Summary:             "ZHtrap 僵尸网络针对 IoT 设备发起扫描与感染，传播 Mirai 变种。",
		SuggestedTechniques: []string{"T1110"},
		SuggestedMalware:    []string{"ZHtrap"},
		SuggestedCves:       []string{"CVE-2021-44228"},
		Tools:               []ToolEntry{{Name: "mirai"}},
		Infrastructure:      []InfraEntry{{Type: "domain", Value: "evil.example"}},
		Malware:             []MalwareEntry{{Name: "ZHtrap", Classification: "botnet"}},
	}
	q := ValidateEnrich(er)
	if !q.OK {
		t.Fatalf("expected pass, hints=%v", q.Hints)
	}
}
