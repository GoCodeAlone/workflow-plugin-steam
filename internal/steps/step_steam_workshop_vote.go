package steps

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_vote — submit a thumbs up/down rating for a Workshop item.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	publishedFileId string (required)
//	vote            string — "up" | "down" | "none" (default: "up")
//	baseUrl         string — override for tests
//
// Outputs: success (bool)
type workshopVoteStep struct{ name string }

// NewWorkshopVoteStep creates a new step.steam_workshop_vote step.
func NewWorkshopVoteStep(name string) sdk.StepInstance {
	return &workshopVoteStep{name: name}
}

func (s *workshopVoteStep) Execute(_ context.Context, _ map[string]any,
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
	steamId, _ := merged["steamId"].(string)
	if steamId == "" {
		return nil, fmt.Errorf("step %s: steamId is required", s.name)
	}
	publishedFileId, _ := merged["publishedFileId"].(string)
	if publishedFileId == "" {
		return nil, fmt.Errorf("step %s: publishedFileId is required", s.name)
	}
	vote, _ := merged["vote"].(string)
	if vote == "" {
		vote = "up"
	}
	baseURL, _ := merged["baseUrl"].(string)

	// Map vote string to Steam API action int.
	// 1 = thumbs up, 2 = thumbs down, 0 = none
	action := "1"
	switch vote {
	case "down":
		action = "2"
	case "none":
		action = "0"
	}

	client := steamclient.New(baseURL)
	params := url.Values{
		"key":             {apiKey},
		"appid":           {appId},
		"steamid":         {steamId},
		"publishedfileid": {publishedFileId},
		"action":          {action},
	}

	_, err := client.Post("/ISteamRemoteStorage/SetUserPublishedFileAction/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: SetUserPublishedFileAction: %w", s.name, err)
	}

	return &sdk.StepResult{Output: map[string]any{
		"success": true,
	}}, nil
}
