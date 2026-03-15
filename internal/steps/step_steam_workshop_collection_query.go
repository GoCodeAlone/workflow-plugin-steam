package steps

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_collection_query — fetch all items in a Steam Workshop collection.
//
// Inputs (current or config):
//
//	apiKey       string (required)
//	appId        string (required)
//	collectionId string (required) — the publishedFileId of the collection item
//	baseUrl      string — override for tests
//
// Outputs: title (string), description (string), items ([]map), itemCount (int)
type workshopCollectionQueryStep struct{ name string }

// NewWorkshopCollectionQueryStep creates a new step.steam_workshop_collection_query step.
func NewWorkshopCollectionQueryStep(name string) sdk.StepInstance {
	return &workshopCollectionQueryStep{name: name}
}

func (s *workshopCollectionQueryStep) Execute(_ context.Context, _ map[string]any,
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
	collectionId, _ := merged["collectionId"].(string)
	if collectionId == "" {
		return nil, fmt.Errorf("step %s: collectionId is required", s.name)
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)
	params := url.Values{
		"key":                   {apiKey},
		"collectioncount":       {"1"},
		"publishedfileids[0]":   {collectionId},
	}

	resp, err := client.Post("/ISteamRemoteStorage/GetCollectionDetails/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: GetCollectionDetails: %w", s.name, err)
	}

	response, _ := resp["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("step %s: unexpected response from GetCollectionDetails", s.name)
	}

	collectionDetails, _ := response["collectiondetails"].([]any)
	if len(collectionDetails) == 0 {
		return &sdk.StepResult{Output: map[string]any{
			"title":       "",
			"description": "",
			"items":       []any{},
			"itemCount":   0,
		}}, nil
	}

	col, _ := collectionDetails[0].(map[string]any)
	title := stringVal(col["title"])
	description := stringVal(col["file_description"])

	children, _ := col["children"].([]any)
	items := make([]any, 0, len(children))
	for _, c := range children {
		child, ok := c.(map[string]any)
		if !ok {
			continue
		}
		items = append(items, map[string]any{
			"publishedFileId": stringVal(child["publishedfileid"]),
		})
	}

	return &sdk.StepResult{Output: map[string]any{
		"title":       title,
		"description": description,
		"items":       items,
		"itemCount":   len(items),
	}}, nil
}
