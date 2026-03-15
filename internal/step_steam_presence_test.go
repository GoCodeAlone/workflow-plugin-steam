package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSteamPresenceSet_Success(t *testing.T) {
	step := newSteamPresenceSetStep("presence")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"steamId": "76561198000000001",
			"status":  "In a game",
			"extra": map[string]any{
				"gameMode": "competitive",
			},
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["applied"] != true {
		t.Errorf("applied = %v, want true", result.Output["applied"])
	}
	keys := result.Output["keys"].([]string)
	if len(keys) < 2 {
		t.Errorf("expected >=2 keys, got %v", keys)
	}
	presence := result.Output["presence"].(map[string]string)
	if presence["status"] != "In a game" {
		t.Errorf("status = %v", presence["status"])
	}
}

func TestSteamPresenceSet_MissingSteamId(t *testing.T) {
	step := newSteamPresenceSetStep("presence")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing steamId")
	}
}

func TestSteamFriendsList_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"friendslist": map[string]any{
				"friends": []any{
					map[string]any{"steamid": "76561198000000002", "relationship": "friend", "friend_since": 1700000000},
					map[string]any{"steamid": "76561198000000003", "relationship": "friend", "friend_since": 1700000001},
				},
			},
		})
	}))
	defer srv.Close()

	step := newSteamFriendsListStep("friends")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "key",
			"steamId": "76561198000000001",
			"baseUrl": srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["count"].(int) != 2 {
		t.Errorf("count = %v, want 2", result.Output["count"])
	}
}

func TestSteamFriendsList_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Private profile — Steam returns {"friendslist": {}} with no friends key
		json.NewEncoder(w).Encode(map[string]any{
			"friendslist": map[string]any{},
		})
	}))
	defer srv.Close()

	step := newSteamFriendsListStep("friends")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "key",
			"steamId": "76561198000000001",
			"baseUrl": srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["count"].(int) != 0 {
		t.Errorf("count = %v, want 0", result.Output["count"])
	}
}

func TestSteamInviteSend_Success(t *testing.T) {
	step := newSteamInviteSendStep("invite")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"steamId": "76561198000000001",
			"lobbyId": "109775241663326011",
			"appId":   "12345",
			"message": "Join my game!",
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inviteURL, _ := result.Output["inviteUrl"].(string)
	if inviteURL == "" {
		t.Error("inviteUrl should not be empty")
	}
	if result.Output["lobbyId"] != "109775241663326011" {
		t.Errorf("lobbyId = %v", result.Output["lobbyId"])
	}
}

func TestSteamInviteSend_MissingLobbyId(t *testing.T) {
	step := newSteamInviteSendStep("invite")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"steamId": "76561198000000001"})
	if err == nil {
		t.Fatal("expected error for missing lobbyId")
	}
}
