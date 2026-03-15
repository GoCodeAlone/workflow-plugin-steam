package workshop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationResult holds the outcome of a workshop package validation.
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Manifest *WorkshopManifest
}

// knownStepTypePrefix lists allowed step type prefixes for strict validation.
// Workshop items may only use safe step types that don't make external calls.
var allowedStrictStepPrefixes = []string{
	"step.set",
	"step.conditional",
	"step.validate",
	"step.game_",
}

// ValidatePackage validates an extracted workshop item directory.
// If strictMode is true, ruleset YAML step types are also checked against
// an allowlist of safe step types.
func ValidatePackage(packageDir string, strictMode bool) (ValidationResult, error) {
	result := ValidationResult{}

	// 1. Read and validate manifest.json.
	manifestPath := filepath.Join(packageDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("missing manifest.json: %v", err))
		return result, nil
	}
	var manifest WorkshopManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("parse manifest.json: %v", err))
		return result, nil
	}
	if err := manifest.Validate(); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}
	result.Manifest = &manifest

	// 2. Validate ruleset.yaml if present.
	rulesetPath := filepath.Join(packageDir, "ruleset.yaml")
	rulesetData, err := os.ReadFile(rulesetPath)
	if err == nil {
		var rulesetDoc map[string]any
		if err := yaml.Unmarshal(rulesetData, &rulesetDoc); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ruleset.yaml: invalid YAML: %v", err))
		} else if strictMode {
			// Check step types against allowlist.
			if steps, ok := rulesetDoc["steps"].([]any); ok {
				for _, s := range steps {
					step, ok := s.(map[string]any)
					if !ok {
						continue
					}
					typeName, _ := step["type"].(string)
					if typeName == "" {
						continue
					}
					if !isAllowedStepType(typeName) {
						result.Errors = append(result.Errors,
							fmt.Sprintf("ruleset.yaml: step type %q is not allowed in Workshop items", typeName))
					}
				}
			}
		}
	}
	// Missing ruleset.yaml is OK for non-ruleset item types.

	result.Valid = len(result.Errors) == 0
	return result, nil
}

// isAllowedStepType returns true if the given step type name is safe for
// Workshop items (matches one of the allowed prefixes).
func isAllowedStepType(typeName string) bool {
	for _, prefix := range allowedStrictStepPrefixes {
		if strings.HasPrefix(typeName, prefix) {
			return true
		}
	}
	return false
}
