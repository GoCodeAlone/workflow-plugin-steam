package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestWorkshopVoteStep_VoteUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	step := steps.NewWorkshopVoteStep("vote_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"publishedFileId": "111",
			"vote":            "up",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["success"] != true {
		t.Errorf("success = %v, want true", result.Output["success"])
	}
}

func TestWorkshopVoteStep_VoteDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	step := steps.NewWorkshopVoteStep("vote_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "s",
			"publishedFileId": "111",
			"vote":            "down",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["success"] != true {
		t.Errorf("success = %v, want true", result.Output["success"])
	}
}

func TestWorkshopVoteStep_MissingPublishedFileId(t *testing.T) {
	step := steps.NewWorkshopVoteStep("vote_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s", "vote": "up"})
	if err == nil {
		t.Fatal("expected error for missing publishedFileId")
	}
}

func TestWorkshopVoteStep_SteamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	step := steps.NewWorkshopVoteStep("vote_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey": "key", "appId": "1", "steamId": "s",
			"publishedFileId": "1", "vote": "up", "baseUrl": srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error from API 403")
	}
}
