package workshop

import "fmt"

// ItemType classifies a Workshop item.
type ItemType string

const (
	ItemTypeRuleset  ItemType = "ruleset"
	ItemTypeCardPack ItemType = "card_pack"
	ItemTypeGameMode ItemType = "game_mode" // ruleset + UI theme
)

// validItemTypes is the set of accepted ItemType values.
var validItemTypes = map[ItemType]bool{
	ItemTypeRuleset:  true,
	ItemTypeCardPack: true,
	ItemTypeGameMode: true,
}

// WorkshopManifest is the deserialized form of a Workshop item's manifest.json.
type WorkshopManifest struct {
	SchemaVersion    int      `json:"schemaVersion"`
	ItemType         ItemType `json:"itemType"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	PreviewImagePath string   `json:"previewImagePath"`
	GameTypes        []string `json:"gameTypes"`
	MinPlayers       int      `json:"minPlayers"`
	MaxPlayers       int      `json:"maxPlayers"`
	Author           string   `json:"author"`  // Steam ID 64 as string
	Version          string   `json:"version"` // semver string
}

// Validate checks that the manifest has required fields and valid values.
func (m *WorkshopManifest) Validate() error {
	if m.Title == "" {
		return fmt.Errorf("workshop manifest: title is required")
	}
	if !validItemTypes[m.ItemType] {
		return fmt.Errorf("workshop manifest: unknown itemType %q (must be ruleset, card_pack, or game_mode)", m.ItemType)
	}
	if len(m.Tags) > 20 {
		return fmt.Errorf("workshop manifest: too many tags (%d); maximum is 20", len(m.Tags))
	}
	return nil
}
