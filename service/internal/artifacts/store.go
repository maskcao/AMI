package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	root string
}

func New(dataDir string) *Store {
	return &Store{root: filepath.Join(dataDir, "store")}
}

func (s *Store) IntelRoot(intelID int64) string {
	return filepath.Join(s.root, "intel", fmt.Sprintf("%d", intelID))
}

func (s *Store) RawPath(intelID int64, seq int) string {
	return filepath.Join(s.IntelRoot(intelID), "raw", fmt.Sprintf("%03d.html", seq))
}

func (s *Store) RawAssetsDir(intelID int64, seq int) string {
	return filepath.Join(s.IntelRoot(intelID), "raw", fmt.Sprintf("%03d", seq), "assets")
}

func (s *Store) WriteRawAsset(intelID int64, seq int, filename string, data []byte) (string, error) {
	dir := s.RawAssetsDir(intelID, seq)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := filepath.Base(filename)
	abs := filepath.Join(dir, name)
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join("store", "intel", fmt.Sprintf("%d", intelID), "raw", fmt.Sprintf("%03d", seq), "assets", name)), nil
}

func (s *Store) ReadRawAsset(intelID int64, seq int, filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.RawAssetsDir(intelID, seq), filepath.Base(filename)))
}

func (s *Store) ParsedDir(intelID int64, version string) string {
	return filepath.Join(s.IntelRoot(intelID), "parsed", version)
}

type ManifestArtifact struct {
	Seq        int    `json:"seq"`
	Role       string `json:"role"`
	Filename   string `json:"filename"`
	Mime       string `json:"mime"`
	SHA256     string `json:"sha256"`
	SizeBytes  int64  `json:"sizeBytes"`
}

type Manifest struct {
	IntelID   int64              `json:"intelId"`
	SourceURL string             `json:"sourceUrl,omitempty"`
	FetchedAt string             `json:"fetchedAt"`
	Artifacts []ManifestArtifact `json:"artifacts"`
}

func (s *Store) WriteRawHTML(intelID int64, seq int, sourceURL string, html []byte) (relPath string, sha string, size int64, err error) {
	dir := filepath.Join(s.IntelRoot(intelID), "raw")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", 0, err
	}
	abs := s.RawPath(intelID, seq)
	if err := os.WriteFile(abs, html, 0o644); err != nil {
		return "", "", 0, err
	}
	sum := sha256.Sum256(html)
	sha = hex.EncodeToString(sum[:])
	size = int64(len(html))
	relPath = filepath.ToSlash(filepath.Join("store", "intel", fmt.Sprintf("%d", intelID), "raw", fmt.Sprintf("%03d.html", seq)))

	m := Manifest{
		IntelID:   intelID,
		SourceURL: sourceURL,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
		Artifacts: []ManifestArtifact{{
			Seq: seq, Role: "html", Filename: fmt.Sprintf("%03d.html", seq),
			Mime: "text/html", SHA256: sha, SizeBytes: size,
		}},
	}
	mb, _ := json.MarshalIndent(m, "", "  ")
	_ = os.WriteFile(filepath.Join(s.IntelRoot(intelID), "manifest.json"), mb, 0o644)
	return relPath, sha, size, nil
}

func (s *Store) WriteParsedMarkdown(intelID int64, version, md string) (mdRel string, err error) {
	dir := s.ParsedDir(intelID, version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	mdAbs := filepath.Join(dir, "body.md")
	if err := os.WriteFile(mdAbs, []byte(md), 0o644); err != nil {
		return "", err
	}
	base := filepath.ToSlash(filepath.Join("store", "intel", fmt.Sprintf("%d", intelID), "parsed", version))
	return base + "/body.md", nil
}

func (s *Store) ReadRaw(intelID int64, seq int) ([]byte, error) {
	return os.ReadFile(s.RawPath(intelID, seq))
}

func (s *Store) ReadParsedText(dataDir string, relPath string) ([]byte, error) {
	if relPath == "" {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(filepath.Join(dataDir, relPath))
}

func (s *Store) FigurePath(intelID int64, version, filename string) string {
	return filepath.Join(s.ParsedDir(intelID, version), "figures", filename)
}

func (s *Store) WriteFigure(intelID int64, version, filename string, data []byte) (relPath string, err error) {
	dir := filepath.Join(s.ParsedDir(intelID, version), "figures")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	abs := filepath.Join(dir, filename)
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join("store", "intel", fmt.Sprintf("%d", intelID), "parsed", version, "figures", filename)), nil
}

func (s *Store) ReadFigure(intelID int64, version, filename string) ([]byte, error) {
	return os.ReadFile(s.FigurePath(intelID, version, filename))
}
