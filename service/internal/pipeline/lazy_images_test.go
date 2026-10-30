package pipeline_test

import (
	"strings"
	"testing"

	"github.com/ami/ami/service/internal/pipeline"
)

func TestHydrateLazyLoadImages(t *testing.T) {
	raw := `<figure><img class="wp-image lozad" data-src="/api/v1/intel/237/raw-assets/ab.png" alt="x" width="100"></figure>`
	out := pipeline.HydrateLazyLoadImages(raw)
	if !strings.Contains(out, `src="/api/v1/intel/237/raw-assets/ab.png"`) {
		t.Fatalf("missing src: %s", out)
	}
	if strings.Contains(out, `src=""`) {
		t.Fatal("empty src")
	}
}

func TestHydrateLazyLoadImages_skipsWhenSrcPresent(t *testing.T) {
	raw := `<img src="/a.png" data-src="/b.png">`
	out := pipeline.HydrateLazyLoadImages(raw)
	if out != raw {
		t.Fatalf("should keep existing src: got %q", out)
	}
}
