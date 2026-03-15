package workshop_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

func TestWorkshopValidator_ValidRuleset(t *testing.T) {
	dir := t.TempDir()
	// Write a valid manifest + ruleset
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	ruleset := "name: test-ruleset\ntriggers: []\nsteps: []\n"
	if err := os.WriteFile(filepath.Join(dir, "ruleset.yaml"), []byte(ruleset), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := workshop.ValidatePackage(dir, false)
	if err != nil {
		t.Fatalf("ValidatePackage: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid=true, got errors: %v", result.Errors)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got: %v", result.Errors)
	}
}

func TestWorkshopValidator_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML (unclosed bracket causes parse error)
	if err := os.WriteFile(filepath.Join(dir, "ruleset.yaml"), []byte("{bad yaml: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := workshop.ValidatePackage(dir, false)
	if err != nil {
		t.Fatalf("ValidatePackage returned unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false for invalid YAML")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one validation error")
	}
}

func TestWorkshopValidator_UnknownStepType(t *testing.T) {
	dir := t.TempDir()
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	// YAML with unknown step type; strictMode=true checks step types
	ruleset := "name: test\nsteps:\n  - name: bad_step\n    type: step.unknown_xyz\n"
	if err := os.WriteFile(filepath.Join(dir, "ruleset.yaml"), []byte(ruleset), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := workshop.ValidatePackage(dir, true)
	if err != nil {
		t.Fatalf("ValidatePackage: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false for unknown step type in strict mode")
	}
}

func TestWorkshopValidator_MissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	// manifest missing title
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"","author":"76561198012345678","version":"1.0.0"}`
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := workshop.ValidatePackage(dir, false)
	if err != nil {
		t.Fatalf("ValidatePackage: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false for manifest with missing title")
	}
}
