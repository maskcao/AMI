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

func TestShouldMergeByTitle_rejectsGenericSite(t *testing.T) {
	if ShouldMergeByTitle("奇安信威胁情报中心") {
		t.Fatal("generic site title must not merge")
	}
	if !ShouldMergeByTitle("金眼狗团伙近期活动跟踪") {
		t.Fatal("real article title should merge")
	}
}

func TestFindKnowledgeID_skipsGenericTitle(t *testing.T) {
	prior := []struct {
		Title        string
		KnowledgeIDs []int64
	}{
		{"奇安信威胁情报中心", []int64{147}},
	}
	if got := FindKnowledgeID("奇安信威胁情报中心", prior); got != 0 {
		t.Fatalf("expected no merge, got %d", got)
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
