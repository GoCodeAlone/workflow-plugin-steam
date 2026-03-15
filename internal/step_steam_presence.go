package internal

import (
	"context"
	"fmt"
	"net/url"

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

	return &sdk.StepResult{Output: map[string]any{
		"friends": friends,
		"count":   len(friends),
	}}, nil
}

// step.steam_invite_send — generates a Steam join-game invite for a lobby.
//
// Steam Web API does not expose a server-side invite endpoint; this step
// constructs a steam:// protocol URL that the host can send via messaging
// (chat, email, etc.) so the invitee can join the lobby directly.
//
// Inputs (current or config):
//
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

	steamId, _ := merged["steamId"].(string)
	if steamId == "" {
		return nil, fmt.Errorf("step %s: steamId is required", s.name)
	}
	lobbyId, _ := merged["lobbyId"].(string)
	if lobbyId == "" {
		return nil, fmt.Errorf("step %s: lobbyId is required", s.name)
	}
	message, _ := merged["message"].(string)

	// steam:// invite URL format for joining a Steam lobby.
	inviteURL := fmt.Sprintf("steam://joinlobby/%s/%s/%s",
		merged["appId"], lobbyId, steamId)
	if appId, _ := merged["appId"].(string); appId == "" {
		inviteURL = fmt.Sprintf("steam://joinlobby/%s/%s", lobbyId, steamId)
	}

	return &sdk.StepResult{Output: map[string]any{
		"inviteUrl": inviteURL,
		"steamId":   steamId,
		"lobbyId":   lobbyId,
		"message":   message,
	}}, nil
}
