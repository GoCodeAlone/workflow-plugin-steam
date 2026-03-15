package internal

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// step.steam_presence_set — sets Rich Presence metadata for a player.
//
// Rich Presence is applied client-side via the Steamworks SDK; this step
// prepares and returns the key/value pairs that the client shell (Tauri)
// should forward to SteamFriends::SetRichPresence().
//
// Inputs (current or config):
//
//	steamId  string (required)
//	status   string — display string (e.g. "In a game")
//	extra    map[string]string — additional key/value pairs
//
// Outputs: steamId, status, keys ([]string), applied (bool)
type steamPresenceSetStep struct{ name string }

func newSteamPresenceSetStep(name string) *steamPresenceSetStep {
	return &steamPresenceSetStep{name: name}
}

func (s *steamPresenceSetStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	steamId, _ := merged["steamId"].(string)
	if steamId == "" {
		return nil, fmt.Errorf("step %s: steamId is required", s.name)
	}
	status, _ := merged["status"].(string)

	keys := []string{}
	presence := map[string]string{}
	if status != "" {
		presence["status"] = status
		keys = append(keys, "status")
	}
	if extra, ok := merged["extra"].(map[string]any); ok {
		for k, v := range extra {
			if sv, ok := v.(string); ok {
				presence[k] = sv
				keys = append(keys, k)
			}
		}
	}

	return &sdk.StepResult{Output: map[string]any{
		"steamId":  steamId,
		"status":   status,
		"keys":     keys,
		"presence": presence,
		"applied":  true,
	}}, nil
}

// step.steam_friends_list — retrieves a player's friends list.
//
// Inputs (current or config):
//
//	apiKey       string (required)
//	steamId      string (required)
//	relationship string — "friend" (default), "all"
//	appId        string — optional; when set, only friends who own this app are returned
//	baseUrl      string — override Steam API base URL (for testing)
//
// Outputs: friends ([]map), count (int)
type steamFriendsListStep struct{ name string }

func newSteamFriendsListStep(name string) *steamFriendsListStep {
	return &steamFriendsListStep{name: name}
}

func (s *steamFriendsListStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	apiKey, _ := merged["apiKey"].(string)
	if apiKey == "" {
		return nil, fmt.Errorf("step %s: apiKey is required", s.name)
	}
	steamId, _ := merged["steamId"].(string)
	if steamId == "" {
		return nil, fmt.Errorf("step %s: steamId is required", s.name)
	}
	relationship, _ := merged["relationship"].(string)
	if relationship == "" {
		relationship = "friend"
	}
	appId, _ := merged["appId"].(string)
	baseURL, _ := merged["baseUrl"].(string)

	client := newSteamClient(baseURL)
	params := url.Values{
		"key":          {apiKey},
		"steamid":      {steamId},
		"relationship": {relationship},
	}

	resp, err := client.get("/ISteamUser/GetFriendList/v1/", params)
	if err != nil {
		return nil, fmt.Errorf("step %s: %w", s.name, err)
	}

	// Shape: {"friendslist": {"friends": [{"steamid": "...", "relationship": "...", "friend_since": N}]}}
	friendslist, _ := resp["friendslist"].(map[string]any)
	friends := []any{}
	if friendslist != nil {
		if f, ok := friendslist["friends"].([]any); ok {
			friends = f
		}
	}

	// If appId is provided, filter friends to those who own the specified app.
	if appId != "" {
		friends = filterFriendsByOwnership(client, apiKey, appId, friends)
	}

	return &sdk.StepResult{Output: map[string]any{
		"friends": friends,
		"count":   len(friends),
	}}, nil
}

// filterFriendsByOwnership returns only the friends who own the given appId.
// Calls IPlayerService/GetOwnedGames/v1/ per friend with appids_filter to minimise
// the data returned; friends with a non-empty game list own the app.
func filterFriendsByOwnership(client *steamClient, apiKey, appId string, friends []any) []any {
	appIdInt, err := strconv.Atoi(appId)
	if err != nil {
		// Non-numeric appId — cannot filter, return all.
		return friends
	}

	filtered := make([]any, 0, len(friends))
	for _, f := range friends {
		fm, ok := f.(map[string]any)
		if !ok {
			continue
		}
		fid, _ := fm["steamid"].(string)
		if fid == "" {
			continue
		}
		params := url.Values{
			"key":              {apiKey},
			"steamid":          {fid},
			"appids_filter[0]": {strconv.Itoa(appIdInt)},
			"include_appinfo":  {"0"},
		}
		resp, err := client.get("/IPlayerService/GetOwnedGames/v1/", params)
		if err != nil {
			// On error (e.g. private profile), skip this friend.
			continue
		}
		response, _ := resp["response"].(map[string]any)
		if response == nil {
			continue
		}
		games, _ := response["games"].([]any)
		if len(games) > 0 {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// step.steam_invite_send — generates a Steam join-game invite for a lobby.
//
// Steam Web API does not expose a server-side invite endpoint; this step
// constructs a steam:// protocol URL that the host can send via messaging
// (chat, email, etc.) so the invitee can join the lobby directly.
//
// Inputs (current or config):
//
//	appId    string (required) — Steam application ID
//	steamId  string (required) — inviting player's Steam ID
//	lobbyId  string (required) — Steam lobby ID to join
//	message  string            — optional custom message
//
// Outputs: inviteUrl, steamId, lobbyId
type steamInviteSendStep struct{ name string }

func newSteamInviteSendStep(name string) *steamInviteSendStep {
	return &steamInviteSendStep{name: name}
}

func (s *steamInviteSendStep) Execute(_ context.Context, _ map[string]any,
	_ map[string]map[string]any, current map[string]any,
	_ map[string]any, config map[string]any) (*sdk.StepResult, error) {

	merged := mergeConfigs(current, config)

	appId, _ := merged["appId"].(string)
	if appId == "" {
		return nil, fmt.Errorf("step %s: appId is required", s.name)
	}
	steamId, _ := merged["steamId"].(string)
	if steamId == "" {
		return nil, fmt.Errorf("step %s: steamId is required", s.name)
	}
	lobbyId, _ := merged["lobbyId"].(string)
	if lobbyId == "" {
		return nil, fmt.Errorf("step %s: lobbyId is required", s.name)
	}
	message, _ := merged["message"].(string)

	// steam:// invite URL format: steam://joinlobby/{appId}/{lobbyId}/{steamId}
	inviteURL := fmt.Sprintf("steam://joinlobby/%s/%s/%s", appId, lobbyId, steamId)

	return &sdk.StepResult{Output: map[string]any{
		"inviteUrl": inviteURL,
		"steamId":   steamId,
		"lobbyId":   lobbyId,
		"message":   message,
	}}, nil
}
