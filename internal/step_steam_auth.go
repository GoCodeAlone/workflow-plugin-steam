package internal

import (
	"context"
	"fmt"
	"net/url"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// step.steam_auth — validates a Steam session ticket via ISteamUserAuth/AuthenticateUserTicket.
//
// Inputs (current or config):
//
//	apiKey  string (required) — Steam Web API key
//	appId   string (required) — Steam App ID
//	ticket  string (required) — hex-encoded session ticket obtained from client-side Steamworks
//	baseUrl string            — override Steam API base URL (for testing)
//
// Outputs: steamId, ownerSteamId, vacBanned, publisherBanned, result
type steamAuthStep struct{ name string }

func newSteamAuthStep(name string) *steamAuthStep { return &steamAuthStep{name: name} }

func (s *steamAuthStep) Execute(_ context.Context, _ map[string]any,
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
	ticket, _ := merged["ticket"].(string)
	if ticket == "" {
		return nil, fmt.Errorf("step %s: ticket is required", s.name)
	}
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)
	params := url.Values{
		"key":    {apiKey},
		"appid":  {appId},
		"ticket": {ticket},
	}

	resp, err := client.get("/ISteamUserAuth/AuthenticateUserTicket/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", s.name, err)
	}

	// Response shape: {"response": {"params": {...}}} on success
	//                 {"response": {"error": {...}}}   on failure
	response, _ := resp["response"].(map[string]any)
	if response == nil {
		return nil, fmt.Errorf("step %s: unexpected Steam API response shape", s.name)
	}

	if errObj, ok := response["error"].(map[string]any); ok {
		errDesc, _ := errObj["errordesc"].(string)
		return nil, fmt.Errorf("step %s: Steam auth error: %s", s.name, errDesc)
	}

	p, _ := response["params"].(map[string]any)
	if p == nil {
		return nil, fmt.Errorf("step %s: Steam auth failed: no params in response", s.name)
	}

	return &sdk.StepResult{Output: map[string]any{
		"steamId":         p["steamid"],
		"ownerSteamId":    p["ownersteamid"],
		"vacBanned":       p["vacbanned"],
		"publisherBanned": p["publisherbanned"],
		"result":          p["result"],
	}}, nil
}
