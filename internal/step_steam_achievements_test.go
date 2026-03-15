package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSteamAchievementSet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	step := newSteamAchievementSetStep("set_ach")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198000000001",
			"achievementName": "FIRST_WIN",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["set"] != true {
		t.Errorf("set = %v, want true", result.Output["set"])
	}
	if result.Output["achievementName"] != "FIRST_WIN" {
		t.Errorf("achievementName = %v, want FIRST_WIN", result.Output["achievementName"])
	}
}

func TestSteamAchievementSet_MissingFields(t *testing.T) {
	step := newSteamAchievementSetStep("set_ach")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s"})
	if err == nil {
		t.Fatal("expected error for missing achievementName")
	}
}

func TestSteamAchievementSync_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"playerstats": map[string]any{
				"steamID": "76561198000000001",
				"gameName": "Test Game",
				"achievements": []any{
					map[string]any{"name": "FIRST_WIN", "achieved": 1},
					map[string]any{"name": "VETERAN", "achieved": 0},
				},
				"stats": []any{
					map[string]any{"name": "wins", "value": 5},
				},
			},
		})
	}))
	defer srv.Close()

	step := newSteamAchievementSyncStep("sync_ach")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "key",
			"appId":   "12345",
			"steamId": "76561198000000001",
			"baseUrl": srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["count"].(int) != 2 {
		t.Errorf("count = %v, want 2", result.Output["count"])
	}
	achs := result.Output["achievements"].([]any)
	if len(achs) != 2 {
		t.Errorf("len(achievements) = %d, want 2", len(achs))
	}
}

func TestSteamStatSet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	step := newSteamStatSetStep("stat_set")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":   "key",
			"appId":    "12345",
			"steamId":  "76561198000000001",
			"statName": "total_wins",
			"value":    float64(42),
			"baseUrl":  srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["statName"] != "total_wins" {
		t.Errorf("statName = %v", result.Output["statName"])
	}
	if result.Output["value"] != float64(42) {
		t.Errorf("value = %v, want 42", result.Output["value"])
	}
}

func TestSteamLeaderboardPush_Success(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/ISteamLeaderboards/FindOrCreateLeaderboard/v2/" {
			json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"leaderboardID": float64(9001),
					"leaderboardName": "GlobalHighScores",
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"leaderboardid":   float64(9001),
					"global_rank_new": float64(3),
					"score":           float64(1500),
				},
			})
		}
	}))
	defer srv.Close()

	step := newSteamLeaderboardPushStep("lb_push")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198000000001",
			"leaderboardName": "GlobalHighScores",
			"score":           float64(1500),
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["leaderboardId"] != "9001" {
		t.Errorf("leaderboardId = %v, want 9001", result.Output["leaderboardId"])
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestSteamLeaderboardPush_MissingFields(t *testing.T) {
	step := newSteamLeaderboardPushStep("lb_push")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s"})
	if err == nil {
		t.Fatal("expected error for missing leaderboardName")
	}
}
