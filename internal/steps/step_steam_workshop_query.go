package steps

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
	steamclient "github.com/GoCodeAlone/workflow-plugin-steam/internal/client"
)

// step.steam_workshop_query — search and paginate Workshop items.
//
// Inputs (current or config):
//
//	apiKey         string (required)
//	appId          string (required)
//	queryType      string — ranked_by_vote | ranked_by_publication_date | ranked_by_trend | text_search | get_by_id
//	searchText     string — used when queryType=text_search
//	tags           string — comma-separated tag filter
//	publishedFileId string — used when queryType=get_by_id
//	page           float64 — 1-based page (default: 1)
//	pageSize       float64 — max 100 (default: 20)
//	baseUrl        string — override for tests
//
// Outputs: items ([]map), totalCount (int), page (int), pageSize (int)
type workshopQueryStep struct{ name string }

// NewWorkshopQueryStep creates a new step.steam_workshop_query step.
func NewWorkshopQueryStep(name string) sdk.StepInstance {
	return &workshopQueryStep{name: name}
}

func (s *workshopQueryStep) Execute(_ context.Context, _ map[string]any,
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

	queryType, _ := merged["queryType"].(string)
	if queryType == "" {
		queryType = "ranked_by_vote"
	}
	searchText, _ := merged["searchText"].(string)
	tags, _ := merged["tags"].(string)
	publishedFileId, _ := merged["publishedFileId"].(string)

	page := 1
	if v, ok := merged["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := 20
	if v, ok := merged["pageSize"].(float64); ok && v > 0 {
		if int(v) > 100 {
			pageSize = 100
		} else {
			pageSize = int(v)
		}
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := steamclient.New(baseURL)

	// Map queryType to Steam API query type int.
	queryTypeInt := workshopQueryTypeInt(queryType)

	params := url.Values{
		"key":               {apiKey},
		"appid":             {appId},
		"query_type":        {strconv.Itoa(queryTypeInt)},
		"page":              {strconv.Itoa(page)},
		"numperpage":        {strconv.Itoa(pageSize)},
		"return_details":    {"1"},
		"return_vote_data":  {"1"},
		"return_tags":       {"1"},
		"return_previews":   {"1"},
	}
	if searchText != "" {
		params.Set("search_text", searchText)
	}
	if tags != "" {
		params.Set("requiredtags[0]", tags)
	}
	if publishedFileId != "" && queryType == "get_by_id" {
		params.Set("publishedfileids[0]", publishedFileId)
	}

	resp, err := client.Get("/IPublishedFileService/QueryFiles/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: QueryFiles: %w", s.name, err)
	}

	response, _ := resp["response"].(map[string]any)
	if response == nil {
		response = map[string]any{}
	}

	totalCount := 0
	if v, ok := response["total"].(float64); ok {
		totalCount = int(v)
	}

	rawItems, _ := response["publishedfiledetails"].([]any)
	if rawItems == nil {
		rawItems = []any{}
	}

	items := make([]any, 0, len(rawItems))
	for _, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		items = append(items, mapWorkshopItem(item))
	}

	return &sdk.StepResult{Output: map[string]any{
		"items":      items,
		"totalCount": totalCount,
		"page":       page,
		"pageSize":   pageSize,
	}}, nil
}

// workshopQueryTypeInt maps queryType string to Steam API integer.
func workshopQueryTypeInt(queryType string) int {
	switch queryType {
	case "ranked_by_vote":
		return 0
	case "ranked_by_publication_date":
		return 1
	case "ranked_by_trend":
		return 3
	case "text_search":
		return 0 // text_search uses the search_text param with ranked_by_vote
	case "get_by_id":
		return 0
	default:
		return 0
	}
}

// mapWorkshopItem converts a raw Steam API item map to our normalized output shape.
func mapWorkshopItem(item map[string]any) map[string]any {
	tags := []string{}
	if rawTags, ok := item["tags"].([]any); ok {
		for _, t := range rawTags {
			if tm, ok := t.(map[string]any); ok {
				if tag, ok := tm["tag"].(string); ok {
					tags = append(tags, tag)
				}
			}
		}
	}

	voteScore := float64(0)
	voteUp := 0
	voteDown := 0
	if vd, ok := item["vote_data"].(map[string]any); ok {
		if v, ok := vd["score"].(float64); ok {
			voteScore = v
		}
		if v, ok := vd["votes_up"].(float64); ok {
			voteUp = int(v)
		}
		if v, ok := vd["votes_down"].(float64); ok {
			voteDown = int(v)
		}
	}

	subscriberCount := 0
	if v, ok := item["subscriptions"].(float64); ok {
		subscriberCount = int(v)
	}

	return map[string]any{
		"publishedFileId": stringVal(item["publishedfileid"]),
		"title":           stringVal(item["title"]),
		"description":     stringVal(item["file_description"]),
		"tags":            tags,
		"previewUrl":      stringVal(item["preview_url"]),
		"subscriberCount": subscriberCount,
		"voteScore":       voteScore,
		"voteUp":          voteUp,
		"voteDown":        voteDown,
		"createdAt":       stringVal(item["time_created"]),
		"updatedAt":       stringVal(item["time_updated"]),
		"authorSteamId":   stringVal(item["creator"]),
	}
}
