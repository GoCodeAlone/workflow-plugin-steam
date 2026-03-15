package workshop_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

func TestWorkshopItemValidate_Valid(t *testing.T) {
	m := &workshop.WorkshopManifest{
		SchemaVersion:    1,
		ItemType:         workshop.ItemTypeRuleset,
		Title:            "Turbo Gwent",
		Description:      "Gwent with 2x speed.",
		Tags:             []string{"gwent", "fast"},
		PreviewImagePath: "assets/preview.png",
		GameTypes:        []string{"gwent"},
		MinPlayers:       2,
		MaxPlayers:       2,
		Author:           "76561198012345678",
		Version:          "1.0.0",
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid manifest failed: %v", err)
	}
}

func TestWorkshopItemValidate_MissingTitle(t *testing.T) {
	m := &workshop.WorkshopManifest{
		SchemaVersion: 1,
		ItemType:      workshop.ItemTypeRuleset,
		Title:         "",
		Author:        "76561198012345678",
		Version:       "1.0.0",
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestWorkshopItemValidate_UnknownType(t *testing.T) {
	m := &workshop.WorkshopManifest{
		SchemaVersion: 1,
		ItemType:      "unknown_type",
		Title:         "Test",
		Author:        "76561198012345678",
		Version:       "1.0.0",
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for unknown item type")
	}
}

func TestWorkshopItemValidate_TooManyTags(t *testing.T) {
	tags := make([]string, 21)
	for i := range tags {
		tags[i] = "tag"
	}
	m := &workshop.WorkshopManifest{
		SchemaVersion: 1,
		ItemType:      workshop.ItemTypeRuleset,
		Title:         "Test",
		Tags:          tags,
		Author:        "76561198012345678",
		Version:       "1.0.0",
	}
	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for >20 tags")
	}
	if !strings.Contains(err.Error(), "tag") {
		t.Errorf("error should mention tags: %v", err)
	}
}

func TestWorkshopItemMarshal_RoundTrip(t *testing.T) {
	m := &workshop.WorkshopManifest{
		SchemaVersion:    1,
		ItemType:         workshop.ItemTypeCardPack,
		Title:            "Starter Pack",
		Description:      "Basic cards.",
		Tags:             []string{"starter"},
		PreviewImagePath: "assets/preview.png",
		GameTypes:        []string{"gwent"},
		MinPlayers:       1,
		MaxPlayers:       4,
		Author:           "76561198012345678",
		Version:          "2.1.0",
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m2 workshop.WorkshopManifest
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m2.Title != m.Title {
		t.Errorf("Title: got %q, want %q", m2.Title, m.Title)
	}
	if m2.ItemType != m.ItemType {
		t.Errorf("ItemType: got %q, want %q", m2.ItemType, m.ItemType)
	}
	if m2.Version != m.Version {
		t.Errorf("Version: got %q, want %q", m2.Version, m.Version)
	}
}
