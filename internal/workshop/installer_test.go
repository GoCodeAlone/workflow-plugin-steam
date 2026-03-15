package workshop_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// makeTestZip creates a valid .steamworkshop zip in memory.
func makeTestZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	w, _ := zw.Create("manifest.json")
	w.Write([]byte(manifest))

	w2, _ := zw.Create("ruleset.yaml")
	w2.Write([]byte("name: test\nsteps: []\n"))

	zw.Close()
	return buf.Bytes()
}

// serveZip creates a temp http server that serves a zip file.
func makeTempZipFile(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.steamworkshop")
	if err != nil {
		t.Fatal(err)
	}
	f.Write(data)
	f.Close()
	return "file://" + f.Name() // installer uses http, but test uses file path
}

func TestInstaller_ExtractZip(t *testing.T) {
	zipData := makeTestZip(t)
	installDir := t.TempDir()

	result, err := workshop.Install(workshop.InstallOptions{
		PublishedFileId: "12345",
		ZipData:         zipData,
		InstallDir:      installDir,
		ValidateRuleset: false,
	})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	// Verify manifest.json was extracted
	if _, err := os.Stat(filepath.Join(result.ItemDir, "manifest.json")); err != nil {
		t.Errorf("manifest.json not found in install dir: %v", err)
	}
	if result.Manifest == nil {
		t.Error("Manifest should not be nil")
	}
	if result.Manifest.Title != "Test" {
		t.Errorf("Manifest.Title = %q, want Test", result.Manifest.Title)
	}
}

func TestInstaller_ValidatesAfterExtract(t *testing.T) {
	// zip with invalid YAML
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	w, _ := zw.Create("manifest.json")
	w.Write([]byte(manifest))
	w2, _ := zw.Create("ruleset.yaml")
	w2.Write([]byte("{bad yaml: [unclosed"))
	zw.Close()

	installDir := t.TempDir()
	_, err := workshop.Install(workshop.InstallOptions{
		PublishedFileId: "12345",
		ZipData:         buf.Bytes(),
		InstallDir:      installDir,
		ValidateRuleset: true,
	})
	if err == nil {
		t.Fatal("expected error for invalid ruleset YAML with validation enabled")
	}
}

func TestInstaller_RejectsInvalidPackage(t *testing.T) {
	// zip missing manifest.json
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("other.yaml")
	w.Write([]byte("name: test\n"))
	zw.Close()

	installDir := t.TempDir()
	_, err := workshop.Install(workshop.InstallOptions{
		PublishedFileId: "12345",
		ZipData:         buf.Bytes(),
		InstallDir:      installDir,
		ValidateRuleset: false,
	})
	if err == nil {
		t.Fatal("expected error for package missing manifest.json")
	}
}

func TestInstaller_IdempotentInstall(t *testing.T) {
	zipData := makeTestZip(t)
	installDir := t.TempDir()

	opts := workshop.InstallOptions{
		PublishedFileId: "12345",
		ZipData:         zipData,
		InstallDir:      installDir,
		ValidateRuleset: false,
	}
	r1, err := workshop.Install(opts)
	if err != nil {
		t.Fatalf("first Install: %v", err)
	}
	if r1.AlreadyCurrent {
		t.Error("first install should not be AlreadyCurrent")
	}

	r2, err := workshop.Install(opts)
	if err != nil {
		t.Fatalf("second Install: %v", err)
	}
	if !r2.AlreadyCurrent {
		t.Error("second install should be AlreadyCurrent (same zip data)")
	}
}
