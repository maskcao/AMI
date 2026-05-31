package pipeline_test

import (
	"path/filepath"
	"testing"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func TestPipelineEndToEnd(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := artifacts.New(dir)

	id, err := db.CreateIntel("Test report", "", "src", "collect")
	if err != nil {
		t.Fatal(err)
	}
	html := "<p>CVE-2024-1234 mentioned here.</p>"
	if err := pipeline.ProcessImport(db, files, id, "http://example.com", []byte(html)); err != nil {
		t.Fatal(err)
	}
	w := pipeline.Worker{DB: db, Files: files}
	for i := 0; i < 40; i++ {
		w.Tick()
	}
	row, err := db.GetIntel(id)
	if err != nil || row == nil {
		t.Fatal("missing intel")
	}
	if row.IngestStatus != "succeeded" {
		t.Fatalf("status=%s", row.IngestStatus)
	}
	if !row.HasParsed || !row.SummarySnippet.Valid {
		t.Fatalf("parsed not reflected: %+v", row)
	}
	manifest := filepath.Join(dir, "store", "intel")
	if row.HasRaw {
		_ = manifest
	}
}
