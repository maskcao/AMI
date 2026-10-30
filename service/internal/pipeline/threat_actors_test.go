package pipeline

import "testing"

func TestExtractThreatActors_aptQ(t *testing.T) {
	text := "Patchwork (APT-Q-36) impersonates university domain names"
	got := ExtractThreatActors(text)
	if len(got) == 0 {
		t.Fatal("expected actors")
	}
	found := false
	for _, a := range got {
		if a == "Patchwork (APT-Q-36)" || a == "APT-Q-36" {
			found = true
		}
	}
	if !found {
		t.Fatalf("got %v", got)
	}
}

func TestExtractThreatActors_cn(t *testing.T) {
	text := "旺刺组织结合 0day 技术开展间谍活动，APT-Q-14 亦被提及"
	got := ExtractThreatActors(text)
	if len(got) < 2 {
		t.Fatalf("got %v want >=2", got)
	}
}

func TestExtractThreatActors_lazarus(t *testing.T) {
	text := "Analysis of Lazarus (APT-Q-1)'s Recent Attacks"
	got := ExtractThreatActors(text)
	if len(got) == 0 {
		t.Fatal("expected lazarus/apt-q-1")
	}
}
