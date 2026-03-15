package steps

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// step.steam_workshop_modcheck — run automated pre-publish moderation checks on a workshop package.
//
// Inputs (current or config):
//
//	packagePath       string   (required) — path to workshop item directory
//	allowedStepTypes  []string            — override default step type allowlist
//
// Outputs: passed (bool), violations ([]string), warnings ([]string)
type workshopModCheckStep struct{ name string }

// NewWorkshopModCheckStep creates a new step.steam_workshop_modcheck step.
func NewWorkshopModCheckStep(name string) sdk.StepInstance {
	return &workshopModCheckStep{name: name}
}

func (s *workshopModCheckStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	packagePath, _ := merged["packagePath"].(string)
	if packagePath == "" {
		return nil, fmt.Errorf("step %s: packagePath is required", s.name)
	}

	// Parse optional allowedStepTypes list.
	var allowedStepTypes []string
	if v, ok := merged["allowedStepTypes"].([]any); ok {
		for _, item := range v {
			if s, ok := item.(string); ok {
				allowedStepTypes = append(allowedStepTypes, s)
			}
		}
	} else if v, ok := merged["allowedStepTypes"].(string); ok && v != "" {
		// Support comma-separated string form.
		for _, part := range strings.Split(v, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				allowedStepTypes = append(allowedStepTypes, trimmed)
			}
		}
	}

	result, err := workshop.RunModCheck(workshop.ModCheckOptions{
		PackageDir:       packagePath,
		AllowedStepTypes: allowedStepTypes,
	})
	if err != nil {
		return nil, fmt.Errorf("step %s: modcheck: %w", s.name, err)
	}

	violationsOut := make([]any, len(result.Violations))
	for i, v := range result.Violations {
		violationsOut[i] = v
	}
	warningsOut := make([]any, len(result.Warnings))
	for i, w := range result.Warnings {
		warningsOut[i] = w
	}

	return &sdk.StepResult{Output: map[string]any{
		"passed":     result.Passed,
		"violations": violationsOut,
		"warnings":   warningsOut,
	}}, nil
}
