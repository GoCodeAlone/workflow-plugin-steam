package workshop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	maxTitleLen       = 128
	maxDescriptionLen = 8000
)

// defaultAllowedStepTypePrefixes is the set of step type prefixes permitted in Workshop items.
var defaultAllowedStepTypePrefixes = []string{
	"step.set",
	"step.conditional",
	"step.validate",
	"step.game_",
}

// allowedAssetExtensions is the set of file extensions permitted in Workshop packages.
var allowedAssetExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".yaml": true,
	".json": true,
	".md":   true,
}

// ModCheckOptions controls the automated moderation check.
type ModCheckOptions struct {
	// PackageDir is the path to the workshop item directory (already extracted).
	PackageDir string
	// AllowedStepTypes overrides the default allowlist. If nil, defaults are used.
	AllowedStepTypes []string
}

// ModCheckResult holds the outcome of the moderation check.
type ModCheckResult struct {
	Passed     bool
	Violations []string
	Warnings   []string
}

// RunModCheck performs automated pre-publish moderation checks on a workshop package directory.
func RunModCheck(opts ModCheckOptions) (ModCheckResult, error) {
	result := ModCheckResult{}

	allowedPrefixes := opts.AllowedStepTypes
	if len(allowedPrefixes) == 0 {
		allowedPrefixes = defaultAllowedStepTypePrefixes
	}

	// 1. Read manifest.json.
	manifestPath := filepath.Join(opts.PackageDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		result.Violations = append(result.Violations, fmt.Sprintf("missing manifest.json: %v", err))
		result.Passed = false
		return result, nil
	}
	var manifest WorkshopManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		result.Violations = append(result.Violations, fmt.Sprintf("parse manifest.json: %v", err))
		result.Passed = false
		return result, nil
	}

	// 2. Field length checks.
	if len(manifest.Title) > maxTitleLen {
		result.Violations = append(result.Violations,
			fmt.Sprintf("title too long (%d chars, max %d)", len(manifest.Title), maxTitleLen))
	}
	if len(manifest.Description) > maxDescriptionLen {
		result.Violations = append(result.Violations,
			fmt.Sprintf("description too long (%d chars, max %d)", len(manifest.Description), maxDescriptionLen))
	}

	// 3. Check all files in the package.
	err = filepath.Walk(opts.PackageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		rel, _ := filepath.Rel(opts.PackageDir, path)

		// Skip modcheck metadata files.
		if info.Name() == ".checksum" {
			return nil
		}

		// Check for executable bit.
		if info.Mode()&0o111 != 0 {
			result.Violations = append(result.Violations,
				fmt.Sprintf("file %q has executable bit set", rel))
		}

		// Check allowed file extensions.
		if ext != "" && !allowedAssetExtensions[ext] {
			result.Violations = append(result.Violations,
				fmt.Sprintf("file %q has disallowed extension %q", rel, ext))
		}

		return nil
	})
	if err != nil {
		return result, fmt.Errorf("modcheck: walk package dir: %w", err)
	}

	// 4. Check ruleset.yaml step types.
	rulesetPath := filepath.Join(opts.PackageDir, "ruleset.yaml")
	if rulesetData, err := os.ReadFile(rulesetPath); err == nil {
		var rulesetDoc map[string]any
		if err := yaml.Unmarshal(rulesetData, &rulesetDoc); err == nil {
			if stepsAny, ok := rulesetDoc["steps"].([]any); ok {
				for _, s := range stepsAny {
					step, ok := s.(map[string]any)
					if !ok {
						continue
					}
					typeName, _ := step["type"].(string)
					if typeName == "" {
						continue
					}
					if !hasAllowedPrefix(typeName, allowedPrefixes) {
						result.Violations = append(result.Violations,
							fmt.Sprintf("step type %q is not in the allowed list for Workshop items", typeName))
					}
				}
			}
		}
	}

	result.Passed = len(result.Violations) == 0
	return result, nil
}

// hasAllowedPrefix checks if a step type name starts with any allowed prefix.
func hasAllowedPrefix(typeName string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(typeName, prefix) {
			return true
		}
	}
	return false
}
