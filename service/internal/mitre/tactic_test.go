package mitre

import "testing"

func TestTacticForTechnique_subtechnique(t *testing.T) {
	if got := TacticForTechnique("T1059.001"); got != "Execution" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveTechniqueTactic_placeholder(t *testing.T) {
	if got := ResolveTechniqueTactic("T1190", "—"); got != "Initial Access" {
		t.Fatalf("got %q", got)
	}
}
