package pipeline

import "testing"

func TestIsImageBytes(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}
	html := []byte("<!doctype html><html>")
	if !IsImageBytes(png) {
		t.Fatal("png")
	}
	if IsImageBytes(html) {
		t.Fatal("html")
	}
}
