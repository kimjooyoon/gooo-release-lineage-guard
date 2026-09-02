package lineage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type InventoryArtifact struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	Digest string `json:"digest"`
}

type Inventory struct {
	Schema               string              `json:"schema"`
	RootReadmeExcluded   bool                `json:"root_readme_excluded"`
	Directories          int                 `json:"directories"`
	RegularFiles         int                 `json:"regular_files"`
	GoFiles              int                 `json:"go_files"`
	GoPhysicalLines      int64               `json:"go_physical_lines"`
	GoooFiles            int                 `json:"gooo_files"`
	GoooPhysicalLines    int64               `json:"gooo_physical_lines"`
	GeneratedArtifacts   []InventoryArtifact `json:"generated_artifacts"`
	WallMS               *int64              `json:"wall_ms"`
	PeakRSSKiB           *int64              `json:"peak_rss_kib"`
}

func BuildInventory(root string) (Inventory, error) {
	result := Inventory{Schema: "gooo/release-lineage-guard/inventory/v1", RootReadmeExcluded: true, WallMS: nil, PeakRSSKiB: nil}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil { return walkErr }
		if path == root { return nil }
		relative, err := filepath.Rel(root, path); if err != nil { return err }
		if strings.HasPrefix(relative, ".git") || strings.HasPrefix(relative, "generated/inventory") { if entry.IsDir() { return fs.SkipDir }; return nil }
		if entry.IsDir() { result.Directories++; return nil }
		if !entry.Type().IsRegular() { return nil }
		result.RegularFiles++
		if relative == "README.md" { result.RegularFiles--; return nil }
		contents, err := os.ReadFile(path); if err != nil { return err }
		lines := int64(strings.Count(string(contents), "\n")); if len(contents) > 0 && !strings.HasSuffix(string(contents), "\n") { lines++ }
		switch filepath.Ext(path) { case ".go": result.GoFiles++; result.GoPhysicalLines += lines; case ".gooo": result.GoooFiles++; result.GoooPhysicalLines += lines }
		if strings.HasPrefix(filepath.ToSlash(relative), "generated/") { digest := sha256.Sum256(contents); result.GeneratedArtifacts = append(result.GeneratedArtifacts, InventoryArtifact{Path: filepath.ToSlash(relative), Bytes: int64(len(contents)), Digest: "sha256:" + hex.EncodeToString(digest[:])}) }
		return nil
	})
	if err != nil { return Inventory{}, fmt.Errorf("inventory %s: %w", root, err) }
	return result, nil
}
