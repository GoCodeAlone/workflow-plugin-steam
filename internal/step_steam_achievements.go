package internal

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// step.steam_achievement_set — unlocks or locks a Steam achievement for a player.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	achievementName string (required) — Steam API name of the achievement
//	value           float64           — 1 to unlock (default), 0 to lock
//	baseUrl         string            — override Steam API base URL (for testing)
//
// Outputs: achievementName, steamId, set (bool)
type steamAchievementSetStep struct{ name string }

func newSteamAchievementSetStep(name string) *steamAchievementSetStep {
	return &steamAchievementSetStep{name: name}
}

func (s *steamAchievementSetStep) Execute(_ context.Context, _ map[string]any,
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
	achievementName, _ := merged["achievementName"].(string)
	if achievementName == "" {
		return nil, fmt.Errorf("step %s: achievementName is required", s.name)
	}
	value := 1
	if v, ok := merged["value"].(float64); ok {
		value = int(v)
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)
	params := url.Values{
		"key":        {apiKey},
		"appid":      {appId},
		"steamid":    {steamId},
		"count":      {"1"},
		"statname[0]": {achievementName},
		"statvalue[0]": {strconv.Itoa(value)},
	}

	_, err := client.post("/ISteamUserStats/SetUserStatsForGame/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", s.name, err)
	}

	return &sdk.StepResult{Output: map[string]any{
		"achievementName": achievementName,
		"steamId":         steamId,
		"set":             true,
	}}, nil
}

// step.steam_achievement_sync — retrieves a player's achievements and stats for an app.
//
// Inputs (current or config):
//
//	apiKey  string (required)
//	appId   string (required)
//	steamId string (required)
//	baseUrl string — override Steam API base URL (for testing)
//
// Outputs: achievements ([]map), stats ([]map), count (int)
type steamAchievementSyncStep struct{ name string }

func newSteamAchievementSyncStep(name string) *steamAchievementSyncStep {
	return &steamAchievementSyncStep{name: name}
}

func (s *steamAchievementSyncStep) Execute(_ context.Context, _ map[string]any,
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
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)
	params := url.Values{
		"key":     {apiKey},
		"appid":   {appId},
		"steamid": {steamId},
	}

	resp, err := client.get("/ISteamUserStats/GetUserStatsForGame/v2/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", s.name, err)
	}

	// Shape: {"playerstats": {"achievements": [...], "stats": [...]}}
	playerStats, _ := resp["playerstats"].(map[string]any)
	if playerStats == nil {
		return &sdk.StepResult{Output: map[string]any{
			"achievements": []any{},
			"stats":        []any{},
			"count":        0,
		}}, nil
	}

	achievements, _ := playerStats["achievements"].([]any)
	if achievements == nil {
		achievements = []any{}
	}
	stats, _ := playerStats["stats"].([]any)
	if stats == nil {
		stats = []any{}
	}

	return &sdk.StepResult{Output: map[string]any{
		"achievements": achievements,
		"stats":        stats,
		"count":        len(achievements),
	}}, nil
}

// step.steam_stat_set — sets a numeric stat for a Steam player.
//
// Inputs (current or config):
//
//	apiKey   string (required)
//	appId    string (required)
//	steamId  string (required)
//	statName string (required) — Steam API name of the stat
//	value    float64 (required)
//	baseUrl  string — override Steam API base URL (for testing)
//
// Outputs: statName, steamId, value, set (bool)
type steamStatSetStep struct{ name string }

func newSteamStatSetStep(name string) *steamStatSetStep { return &steamStatSetStep{name: name} }

func (s *steamStatSetStep) Execute(_ context.Context, _ map[string]any,
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
	statName, _ := merged["statName"].(string)
	if statName == "" {
		return nil, fmt.Errorf("step %s: statName is required", s.name)
	}
	valueF, _ := merged["value"].(float64)
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)
	params := url.Values{
		"key":          {apiKey},
		"appid":        {appId},
		"steamid":      {steamId},
		"count":        {"1"},
		"statname[0]":  {statName},
		"statvalue[0]": {strconv.FormatFloat(valueF, 'f', -1, 64)},
	}

	_, err := client.post("/ISteamUserStats/SetUserStatsForGame/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", s.name, err)
	}

	return &sdk.StepResult{Output: map[string]any{
		"statName": statName,
		"steamId":  steamId,
		"value":    valueF,
		"set":      true,
	}}, nil
}

// step.steam_leaderboard_push — finds or creates a leaderboard and uploads a score.
//
// Inputs (current or config):
//
//	apiKey          string (required)
//	appId           string (required)
//	steamId         string (required)
//	leaderboardName string (required)
//	score           float64 (required)
//	scoreMethod     string  — "KeepBest" (default), "ForceUpdate"
//	sortMethod      string  — "Descending" (default), "Ascending"
//	displayType     string  — "Numeric" (default), "TimeSeconds", "TimeMilliSeconds"
//	baseUrl         string  — override Steam API base URL (for testing)
//
// Outputs: leaderboardId, steamId, score, globalRank
type steamLeaderboardPushStep struct{ name string }

func newSteamLeaderboardPushStep(name string) *steamLeaderboardPushStep {
	return &steamLeaderboardPushStep{name: name}
}

func (s *steamLeaderboardPushStep) Execute(_ context.Context, _ map[string]any,
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
	leaderboardName, _ := merged["leaderboardName"].(string)
	if leaderboardName == "" {
		return nil, fmt.Errorf("step %s: leaderboardName is required", s.name)
	}
	scoreF, _ := merged["score"].(float64)
	scoreMethod, _ := merged["scoreMethod"].(string)
	if scoreMethod == "" {
		scoreMethod = "KeepBest"
	}
	sortMethod, _ := merged["sortMethod"].(string)
	if sortMethod == "" {
		sortMethod = "Descending"
	}
	displayType, _ := merged["displayType"].(string)
	if displayType == "" {
		displayType = "Numeric"
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)

	// Step 1: find or create the leaderboard.
	findParams := url.Values{
		"key":              {apiKey},
		"appid":            {appId},
		"name":             {leaderboardName},
		"sortmethod":       {sortMethod},
		"displaytype":      {displayType},
		"createifnotfound": {"1"},
	}
	findResp, err := client.post("/ISteamLeaderboards/FindOrCreateLeaderboard/v2/", findParams)
	if err != nil {
		return nil, fmt.Errorf("step %s: find/create leaderboard: %w", s.name, err)
	}

	// Shape: {"result": {"leaderboardID": N, ...}}
	resultObj, _ := findResp["result"].(map[string]any)
	if resultObj == nil {
		return nil, fmt.Errorf("step %s: find/create leaderboard: unexpected response shape", s.name)
	}
	leaderboardID := ""
	switch v := resultObj["leaderboardID"].(type) {
	case float64:
		leaderboardID = strconv.FormatFloat(v, 'f', 0, 64)
	case string:
		leaderboardID = v
	}
	if leaderboardID == "" {
		return nil, fmt.Errorf("step %s: no leaderboardID in response", s.name)
	}

	// Step 2: upload the score.
	uploadParams := url.Values{
		"key":           {apiKey},
		"appid":         {appId},
		"steamid":       {steamId},
		"leaderboardid": {leaderboardID},
		"score":         {strconv.FormatFloat(scoreF, 'f', 0, 64)},
		"scoremethod":   {scoreMethod},
	}
	uploadResp, err := client.post("/ISteamLeaderboards/UploadLeaderboardScore/v1/", uploadParams)
	if err != nil {
		return nil, fmt.Errorf("step %s: upload score: %w", s.name, err)
	}

	// Shape: {"result": {"global_rank_new": N, ...}}
	var globalRank any
	if r, ok := uploadResp["result"].(map[string]any); ok {
		globalRank = r["global_rank_new"]
	}

	return &sdk.StepResult{Output: map[string]any{
		"leaderboardId": leaderboardID,
		"steamId":       steamId,
		"score":         scoreF,
		"globalRank":    globalRank,
	}}, nil
}
