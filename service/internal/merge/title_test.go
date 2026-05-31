package merge

import "testing"

func TestTitlesMatch(t *testing.T) {
	if !TitlesMatch("  foo   bar ", "foo bar") {
		t.Fatal("expected match")
	}
	if TitlesMatch("Foo", "foo") {
		t.Fatal("case sensitive")
	}
}

func TestFindKnowledgeID(t *testing.T) {
	prior := []struct {
		Title        string
		KnowledgeIDs []int64
	}{
		{"Other", []int64{9}},
		{"Same Title", []int64{1001}},
	}
	if got := FindKnowledgeID("Same  Title", prior); got != 1001 {
		t.Fatalf("got %d want 1001", got)
	}
	if got := FindKnowledgeID("New", prior); got != 0 {
		t.Fatalf("got %d want 0", got)
	}
}
