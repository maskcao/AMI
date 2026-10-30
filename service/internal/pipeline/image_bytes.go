package pipeline

import (
	"bytes"
	"strings"
)

// IsImageBytes reports whether body looks like a supported raster image (not HTML/text).
func IsImageBytes(body []byte) bool {
	if len(body) < 12 {
		return false
	}
	if bytes.HasPrefix(body, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}) {
		return true
	}
	if body[0] == 0xff && body[1] == 0xd8 && body[2] == 0xff {
		return true
	}
	if bytes.HasPrefix(body, []byte("GIF87a")) || bytes.HasPrefix(body, []byte("GIF89a")) {
		return true
	}
	if len(body) >= 12 && string(body[0:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return true
	}
	low := strings.ToLower(string(body[:minInt(len(body), 256)]))
	if strings.HasPrefix(low, "<!doctype") || strings.HasPrefix(low, "<html") {
		return false
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
