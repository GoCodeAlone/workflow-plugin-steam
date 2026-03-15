package steps

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_unsubscribe — unsubscribe from a Workshop item and optionally remove its files.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	publishedFileId string (required)
//	installDir      string — base install directory (default: "data/workshop")
//	removeFiles     bool   — delete installed item dir (default: true)
//	baseUrl         string — override for tests
//
// Outputs: unsubscribed (bool), filesRemoved (bool)
type workshopUnsubscribeStep struct{ name string }

// NewWorkshopUnsubscribeStep creates a new step.steam_workshop_unsubscribe step.
func NewWorkshopUnsubscribeStep(name string) sdk.StepInstance {
	return &workshopUnsubscribeStep{name: name}
}

func (s *workshopUnsubscribeStep) Execute(_ context.Context, _ map[string]any,
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

	installDir, _ := merged["installDir"].(string)
	if installDir == "" {
		installDir = "data/workshop"
	}

	removeFiles := true
	if v, ok := merged["removeFiles"].(bool); ok {
		removeFiles = v
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)
	params := url.Values{
		"key":             {apiKey},
		"appid":           {appId},
		"steamid":         {steamId},
		"publishedfileid": {publishedFileId},
	}

	_, err := client.Post("/ISteamRemoteStorage/UnsubscribePublishedFile/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: UnsubscribePublishedFile: %w", s.name, err)
	}

	filesRemoved := false
	if removeFiles {
		itemDir := filepath.Join(installDir, publishedFileId)
		if err := os.RemoveAll(itemDir); err != nil {
			return nil, fmt.Errorf("step %s: remove item dir %q: %w", s.name, itemDir, err)
		}
		filesRemoved = true
	}

	return &sdk.StepResult{Output: map[string]any{
		"unsubscribed": true,
		"filesRemoved": filesRemoved,
	}}, nil
}
