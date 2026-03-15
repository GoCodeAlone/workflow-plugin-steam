package steps

import (
	"context"
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// step.steam_workshop_validate — validates an extracted Workshop package directory.
//
// Inputs (current or config):
//
//	packagePath string (required) — path to extracted workshop zip directory
//	strictMode  bool              — if true, validates step types in ruleset YAML
//
// Outputs: valid (bool), errors ([]string), manifest (map)
type workshopValidateStep struct{ name string }

// NewWorkshopValidateStep creates a new step.steam_workshop_validate step.
func NewWorkshopValidateStep(name string) sdk.StepInstance {
	return &workshopValidateStep{name: name}
}

func (s *workshopValidateStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	packagePath, _ := merged["packagePath"].(string)
	if packagePath == "" {
		return nil, fmt.Errorf("step %s: packagePath is required", s.name)
	}
	strictMode := false
	if v, ok := merged["strictMode"].(bool); ok {
		strictMode = v
	}

	result, err := workshop.ValidatePackage(packagePath, strictMode)
	if err != nil {
		return nil, fmt.Errorf("step %s: validate package: %w", s.name, err)
	}

	// Convert errors slice to []any for output map.
	errorsOut := make([]any, len(result.Errors))
	for i, e := range result.Errors {
		errorsOut[i] = e
	}

	manifestOut := map[string]any{}
	if result.Manifest != nil {
		m := result.Manifest
		manifestOut = map[string]any{
			"schemaVersion":    m.SchemaVersion,
			"itemType":         string(m.ItemType),
			"title":            m.Title,
			"description":      m.Description,
			"tags":             m.Tags,
			"previewImagePath": m.PreviewImagePath,
			"gameTypes":        m.GameTypes,
			"minPlayers":       m.MinPlayers,
			"maxPlayers":       m.MaxPlayers,
			"author":           m.Author,
			"version":          m.Version,
		}
	}

	return &sdk.StepResult{Output: map[string]any{
		"valid":    result.Valid,
		"errors":   errorsOut,
		"manifest": manifestOut,
	}}, nil
}

// mergeConfigs merges current (runtime context) and config (step config),
// with current taking precedence.
func mergeConfigs(current, config map[string]any) map[string]any {
	merged := make(map[string]any, len(config)+len(current))
	for k, v := range config {
		merged[k] = v
	}
	for k, v := range current {
		merged[k] = v
	}
	return merged
}
