package steps

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_publish — create or update a Steam Workshop item via ISteamRemoteStorage.
//
// Inputs (current or config):
//
//	apiKey          string (required) — Steam Web API key
//	appId           string (required) — Steam App ID
//	steamId         string (required) — publisher's Steam ID 64
//	packagePath     string (required) — path to .steamworkshop zip file
//	title           string (required on first publish)
//	description     string
//	tags            string — comma-separated tag list
//	previewImagePath string — local path to preview image
//	publishedFileId string — if set, update existing item
//	visibility      string — "public" | "friendsonly" | "private" (default: "private")
//	changelog       string — update notes for this version
//	baseUrl         string — override for tests
//
// Outputs: publishedFileId (string), itemUrl (string), created (bool)
type workshopPublishStep struct{ name string }

// NewWorkshopPublishStep creates a new step.steam_workshop_publish step.
func NewWorkshopPublishStep(name string) sdk.StepInstance {
	return &workshopPublishStep{name: name}
}

func (s *workshopPublishStep) Execute(_ context.Context, _ map[string]any,
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
	packagePath, _ := merged["packagePath"].(string)
	if packagePath == "" {
		return nil, fmt.Errorf("step %s: packagePath is required", s.name)
	}

	title, _ := merged["title"].(string)
	description, _ := merged["description"].(string)
	tags, _ := merged["tags"].(string)
	previewImagePath, _ := merged["previewImagePath"].(string)
	publishedFileId, _ := merged["publishedFileId"].(string)
	visibility, _ := merged["visibility"].(string)
	if visibility == "" {
		visibility = "private"
	}
	changelog, _ := merged["changelog"].(string)
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)

	var resultFileId string
	created := false

	if publishedFileId == "" {
		// New item: call PublishWorkshopFile
		if title == "" {
			return nil, fmt.Errorf("step %s: title is required for new Workshop items", s.name)
		}
		params := url.Values{
			"key":             {apiKey},
			"appid":           {appId},
			"steamid":         {steamId},
			"ugctype":         {"0"}, // 0 = items, readytouseitem
			"title":           {title},
			"gamedescription": {description},
			"gametype":        {tags},
			"visibility":      {visibilityInt(visibility)},
		}
		if previewImagePath != "" {
			params.Set("previewfilepath", previewImagePath)
		}
		resp, err := client.Post("/ISteamRemoteStorage/PublishWorkshopFile/v1/", params)
		if err != nil {
			return nil, fmt.Errorf("step %s: PublishWorkshopFile: %w", s.name, err)
		}
		response, _ := resp["response"].(map[string]any)
		if response == nil {
			return nil, fmt.Errorf("step %s: unexpected response from PublishWorkshopFile", s.name)
		}
		resultFileId = stringVal(response["publishedfileid"])
		if resultFileId == "" {
			resultFileId = publishedFileId
		}
		created = true
	} else {
		// Update existing item: call UpdatePublishedFileDetails
		params := url.Values{
			"key":             {apiKey},
			"appid":           {appId},
			"steamid":         {steamId},
			"publishedfileid": {publishedFileId},
			"visibility":      {visibilityInt(visibility)},
		}
		if title != "" {
			params.Set("title", title)
		}
		if description != "" {
			params.Set("file_description", description)
		}
		if tags != "" {
			params.Set("tags[0]", tags)
		}
		if changelog != "" {
			params.Set("change_description", changelog)
		}
		_, err := client.Post("/ISteamRemoteStorage/UpdatePublishedFileDetails/v1/", params)
		if err != nil {
			return nil, fmt.Errorf("step %s: UpdatePublishedFileDetails: %w", s.name, err)
		}
		resultFileId = publishedFileId
		created = false
	}

	itemUrl := ""
	if resultFileId != "" {
		itemUrl = "https://steamcommunity.com/sharedfiles/filedetails/?id=" + resultFileId
	}

	return &sdk.StepResult{Output: map[string]any{
		"publishedFileId": resultFileId,
		"itemUrl":         itemUrl,
		"created":         created,
	}}, nil
}

// visibilityInt converts a visibility string to Steam API integer string.
func visibilityInt(v string) string {
	switch v {
	case "public":
		return "0"
	case "friendsonly":
		return "1"
	default: // "private"
		return "2"
	}
}

// stringVal extracts a string from any, handling both string and float64 types.
func stringVal(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%.0f", x)
	}
	return ""
}
