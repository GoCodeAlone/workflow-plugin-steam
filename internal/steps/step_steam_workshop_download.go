package steps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

// step.steam_workshop_download — subscribe, download, extract, and validate a Workshop item.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	publishedFileId string (required)
//	installDir      string — target base directory (default: "data/workshop")
//	subscribeOnly   bool   — record subscription but skip download (default: false)
//	validateRuleset bool   — run server-side validation after install (default: true)
//	baseUrl         string — override for tests
//
// Outputs: installed (bool), itemDir (string), manifest (map), alreadyCurrent (bool)
type workshopDownloadStep struct{ name string }

// NewWorkshopDownloadStep creates a new step.steam_workshop_download step.
func NewWorkshopDownloadStep(name string) sdk.StepInstance {
	return &workshopDownloadStep{name: name}
}

func (s *workshopDownloadStep) Execute(_ context.Context, _ map[string]any,
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

	subscribeOnly := false
	if v, ok := merged["subscribeOnly"].(bool); ok {
		subscribeOnly = v
	}

	validateRuleset := true
	if v, ok := merged["validateRuleset"].(bool); ok {
		validateRuleset = v
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)

	// 1. Subscribe to the item (server-side subscription tracking).
	subParams := url.Values{
		"key":             {apiKey},
		"appid":           {appId},
		"steamid":         {steamId},
		"publishedfileid": {publishedFileId},
	}
	if _, err := client.Post("/ISteamRemoteStorage/SubscribePublishedFile/v1/", subParams); err != nil {
		return nil, fmt.Errorf("step %s: SubscribePublishedFile: %w", s.name, err)
	}

	if subscribeOnly {
		return &sdk.StepResult{Output: map[string]any{
			"installed":      false,
			"itemDir":        "",
			"manifest":       map[string]any{},
			"alreadyCurrent": false,
		}}, nil
	}

	// 2. Get file details to retrieve the download URL.
	detailsParams := url.Values{
		"key":               {apiKey},
		"itemcount":         {"1"},
		"publishedfileids[0]": {publishedFileId},
	}
	detailsResp, err := client.Post("/ISteamRemoteStorage/GetPublishedFileDetails/v1/", detailsParams)
	if err != nil {
		return nil, fmt.Errorf("step %s: GetPublishedFileDetails: %w", s.name, err)
	}

	response, _ := detailsResp["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("step %s: unexpected response from GetPublishedFileDetails", s.name)
	}
	details, _ := response["publishedfiledetails"].([]any)
	if len(details) == 0 {
		return nil, fmt.Errorf("step %s: no file details in response", s.name)
	}
	fileDetail, _ := details[0].(map[string]any)
	fileURL, _ := fileDetail["file_url"].(string)
	if fileURL == "" {
		return nil, fmt.Errorf("step %s: no file_url in Workshop item details", s.name)
	}

	// 3. Download the zip.
	zipData, err := downloadFile(fileURL)
	if err != nil {
		return nil, fmt.Errorf("step %s: download Workshop zip: %w", s.name, err)
	}

	// 4. Install (extract + optional validate).
	installResult, err := workshop.Install(workshop.InstallOptions{
		PublishedFileId: publishedFileId,
		ZipData:         zipData,
		InstallDir:      installDir,
		ValidateRuleset: validateRuleset,
	})
	if err != nil {
		return nil, fmt.Errorf("step %s: install: %w", s.name, err)
	}

	manifestOut := map[string]any{}
	if installResult.Manifest != nil {
		m := installResult.Manifest
		manifestOut = map[string]any{
			"itemType": string(m.ItemType),
			"title":    m.Title,
			"version":  m.Version,
			"author":   m.Author,
		}
	}

	return &sdk.StepResult{Output: map[string]any{
		"installed":      true,
		"itemDir":        installResult.ItemDir,
		"manifest":       manifestOut,
		"alreadyCurrent": installResult.AlreadyCurrent,
	}}, nil
}

// downloadFile fetches a URL and returns the response body bytes.
func downloadFile(fileURL string) ([]byte, error) {
	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", fileURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", fileURL, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
