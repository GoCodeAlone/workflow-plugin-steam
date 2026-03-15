package workshop_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// makeTestWorkshopDir creates a minimal workshop source directory on disk.
func makeTestWorkshopDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test Ruleset","description":"Desc","tags":["test"],"previewImagePath":"assets/preview.png","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "ruleset.yaml"), []byte("name: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Small valid PNG-like content (not a real PNG but enough to test packaging)
	if err := os.WriteFile(filepath.Join(assetsDir, "preview.png"), []byte("PNGDATA"), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestPackager_BuildsZip(t *testing.T) {
	srcDir := makeTestWorkshopDir(t)
	outPath := filepath.Join(t.TempDir(), "output.steamworkshop")

	if err := workshop.BuildPackage(workshop.PackageOptions{
		SourceDir:  srcDir,
		OutputPath: outPath,
	}); err != nil {
		t.Fatalf("BuildPackage: %v", err)
	}

	// Verify it's a valid zip containing manifest.json and ruleset.yaml
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output zip: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip reader: %v", err)
	}

	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["manifest.json"] {
		t.Error("zip missing manifest.json")
	}
	if !names["ruleset.yaml"] {
		t.Error("zip missing ruleset.yaml")
	}
}

func TestPackager_RejectsInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	// manifest with no title
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"","author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "output.steamworkshop")
	err := workshop.BuildPackage(workshop.PackageOptions{
		SourceDir:  dir,
		OutputPath: outPath,
	})
	if err == nil {
		t.Fatal("expected error for invalid manifest (empty title)")
	}
}

func TestPackager_RejectsLargeAssets(t *testing.T) {
	srcDir := makeTestWorkshopDir(t)
	// Overwrite preview with > 1MB content
	bigContent := make([]byte, 1024*1024+1)
	if err := os.WriteFile(filepath.Join(srcDir, "assets", "preview.png"), bigContent, 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "output.steamworkshop")
	err := workshop.BuildPackage(workshop.PackageOptions{
		SourceDir:       srcDir,
		OutputPath:      outPath,
		MaxPreviewBytes: 1024 * 1024,
	})
	if err == nil {
		t.Fatal("expected error for preview image > 1MB")
	}
}

func TestPackager_MaxSizeEnforced(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Big","description":"","tags":[],"previewImagePath":"","gameTypes":[],"minPlayers":1,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a file that exceeds the total max
	bigFile := make([]byte, 200)
	if err := os.WriteFile(filepath.Join(dir, "big.yaml"), bigFile, 0o644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(t.TempDir(), "output.steamworkshop")
	err := workshop.BuildPackage(workshop.PackageOptions{
		SourceDir:     dir,
		OutputPath:    outPath,
		MaxTotalBytes: 100, // tiny limit
	})
	if err == nil {
		t.Fatal("expected error for total size > MaxTotalBytes")
	}
}
