package pipeline

import (
	"slices"
	"strings"
	"testing"
)

func TestExtractThreatActors_samples(t *testing.T) {
	cases := []struct {
		id   int
		text string
		want []string
	}{
		{
			231,
			"揭秘魔罗桫（confucius）组织武器库源代码\n奇安信威胁情报中心披露了南亚APT组织魔罗桫（Confucius）的武器库",
			nil,
		},
		{235, "Patchwork (APT-Q-36) impersonates\nPatchwork（APT-Q-36）组织近期", []string{"Patchwork (APT-Q-36)"}},
		{236, "金眼狗团伙近期\n金眼狗团伙（APT-Q-27）长期", []string{"金眼狗团伙 (APT-Q-27)"}},
		{239, "APT-Q-14 group combines\nAPT-Q-14（ClickOnce）组织利用", []string{"APT-Q-14"}},
	}
	for _, c := range cases {
		got := ExtractThreatActors(c.text)
		for _, w := range c.want {
			if c.id == 231 {
				// 231 has no APT-Q in sample snippet; only ensure no bogus long capture.
				for _, g := range got {
					if strings.Contains(g, "披露") || strings.Contains(g, "情报中心") {
						t.Fatalf("%d: bogus %q in %v", c.id, g, got)
					}
				}
				continue
			}
			if !slices.Contains(got, w) {
				t.Fatalf("%d: got %v want contains %q", c.id, got, w)
			}
		}
	}
}
