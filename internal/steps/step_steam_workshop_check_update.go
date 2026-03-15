package steps

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// step.steam_workshop_check_update — check subscribed items for available updates.
//
// Inputs (current or config):
//
//	apiKey           string   (required)
//	appId            string   (required)
//	installDir       string   — base install directory (default: "data/workshop")
//	publishedFileIds []string — specific IDs to check (default: reads all from versiondb)
//	baseUrl          string   — override for tests
//
// Outputs: updates ([]map{publishedFileId, currentUpdatedAt, newUpdatedAt, hasUpdate})
type workshopCheckUpdateStep struct{ name string }

// NewWorkshopCheckUpdateStep creates a new step.steam_workshop_check_update step.
func NewWorkshopCheckUpdateStep(name string) sdk.StepInstance {
	return &workshopCheckUpdateStep{name: name}
}

func (s *workshopCheckUpdateStep) Execute(_ context.Context, _ map[string]any,
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
	installDir, _ := merged["installDir"].(string)
	if installDir == "" {
		installDir = "data/workshop"
	}
	baseURL, _ := merged["baseUrl"].(string)

	// Load versiondb.
	db := workshop.NewVersionDB(filepath.Join(installDir, ".versions.json"))

	// Collect IDs to check.
	var fileIds []string
	if rawIds, ok := merged["publishedFileIds"].([]any); ok {
		for _, id := range rawIds {
			if s, ok := id.(string); ok && s != "" {
				fileIds = append(fileIds, s)
			}
		}
	}
	if len(fileIds) == 0 {
		// Read all from versiondb.
		records, err := db.ListAll()
		if err != nil {
			return nil, fmt.Errorf("step %s: list versiondb: %w", s.name, err)
		}
		for _, r := range records {
			fileIds = append(fileIds, r.PublishedFileId)
		}
	}

	if len(fileIds) == 0 {
		return &sdk.StepResult{Output: map[string]any{"updates": []any{}}}, nil
	}

	// Call GetPublishedFileDetails for all IDs.
	client := steamclient.New(baseURL)
	params := url.Values{
		"key":       {apiKey},
		"itemcount": {fmt.Sprintf("%d", len(fileIds))},
	}
	for i, id := range fileIds {
		params.Set(fmt.Sprintf("publishedfileids[%d]", i), id)
	}

	resp, err := client.Post("/ISteamRemoteStorage/GetPublishedFileDetails/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: GetPublishedFileDetails: %w", s.name, err)
	}

	response, _ := resp["response"].(map[string]any)
	if response == nil {
		response = map[string]any{}
	}
	rawDetails, _ := response["publishedfiledetails"].([]any)

	updates := make([]any, 0, len(rawDetails))
	for _, raw := range rawDetails {
		detail, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		fileId := stringVal(detail["publishedfileid"])
		apiUpdatedUnix := int64(0)
		if v, ok := detail["time_updated"].(float64); ok {
			apiUpdatedUnix = int64(v)
		}
		apiUpdatedAt := time.Unix(apiUpdatedUnix, 0).UTC()

		// Look up current installed version.
		record, found := db.Get(fileId)
		hasUpdate := !found || apiUpdatedAt.After(record.LastUpdatedAt)

		currentUpdatedStr := ""
		if found {
			currentUpdatedStr = record.LastUpdatedAt.Format(time.RFC3339)
		}

		updates = append(updates, map[string]any{
			"publishedFileId":   fileId,
			"currentUpdatedAt":  currentUpdatedStr,
			"newUpdatedAt":      apiUpdatedAt.Format(time.RFC3339),
			"hasUpdate":         hasUpdate,
		})
	}

	return &sdk.StepResult{Output: map[string]any{
		"updates": updates,
	}}, nil
}
