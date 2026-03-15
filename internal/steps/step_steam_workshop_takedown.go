package steps

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_takedown — admin-only step to remove a Workshop item via publisher API key.
//
// Inputs (current or config):
//
//	apiKey          string (required) — publisher Steam Web API key (server-side only)
//	appId           string (required)
//	publishedFileId string (required)
//	reason          string — recorded in audit log
//	baseUrl         string — override for tests
//
// Outputs: removed (bool)
type workshopTakedownStep struct{ name string }

// NewWorkshopTakedownStep creates a new step.steam_workshop_takedown step.
func NewWorkshopTakedownStep(name string) sdk.StepInstance {
	return &workshopTakedownStep{name: name}
}

func (s *workshopTakedownStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	apiKey, _ := merged["apiKey"].(string)
	if apiKey == "" {
		return nil, fmt.Errorf("step %s: apiKey is required", s.name)
	}
	appId, _ := merged["appId"].(string)
	if appId == "" {
		return nil, fmt.Errorf("step %s: appId is required", s.name)
	}
	publishedFileId, _ := merged["publishedFileId"].(string)
	if publishedFileId == "" {
		return nil, fmt.Errorf("step %s: publishedFileId is required", s.name)
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)
	params := url.Values{
		"key":             {apiKey},
		"appid":           {appId},
		"publishedfileid": {publishedFileId},
	}

	_, err := client.Post("/ISteamRemoteStorage/DeletePublishedFile/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: DeletePublishedFile: %w", s.name, err)
	}

	return &sdk.StepResult{Output: map[string]any{
		"removed": true,
	}}, nil
}
