package steps

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_report — submit a moderation report to Steam.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	publishedFileId string (required)
//	reportType      string (required) — "legal" | "harassment" | "spam" | "other"
//	reportText      string — free text description
//	baseUrl         string — override for tests
//
// Outputs: reported (bool)
type workshopReportStep struct{ name string }

// NewWorkshopReportStep creates a new step.steam_workshop_report step.
func NewWorkshopReportStep(name string) sdk.StepInstance {
	return &workshopReportStep{name: name}
}

var validReportTypes = map[string]string{
	"legal":       "1",
	"harassment":  "2",
	"spam":        "3",
	"other":       "4",
}

func (s *workshopReportStep) Execute(_ context.Context, _ map[string]any,
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
	reportType, _ := merged["reportType"].(string)
	reportTypeInt, ok := validReportTypes[reportType]
	if !ok {
		return nil, fmt.Errorf("step %s: invalid reportType %q (must be legal, harassment, spam, or other)", s.name, reportType)
	}
	reportText, _ := merged["reportText"].(string)
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)
	params := url.Values{
		"key":             {apiKey},
		"appid":           {appId},
		"steamid":         {steamId},
		"publishedfileid": {publishedFileId},
		"reporttype":      {reportTypeInt},
	}
	if reportText != "" {
		params.Set("reporttext", reportText)
	}

	_, err := client.Post("/ISteamRemoteStorage/ReportContent/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: ReportContent: %w", s.name, err)
	}

	return &sdk.StepResult{Output: map[string]any{
		"reported": true,
	}}, nil
}
