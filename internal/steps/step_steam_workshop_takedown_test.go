package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestWorkshopTakedownStep_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	step := steps.NewWorkshopTakedownStep("takedown_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "publisherkey",
			"appId":           "12345",
			"publishedFileId": "999",
			"reason":          "policy_violation",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["removed"] != true {
		t.Errorf("removed = %v, want true", result.Output["removed"])
	}
}

func TestWorkshopTakedownStep_MissingApiKey(t *testing.T) {
	step := steps.NewWorkshopTakedownStep("takedown_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"appId": "1", "publishedFileId": "1"})
	if err == nil {
		t.Fatal("expected error for missing apiKey")
	}
}

func TestWorkshopTakedownStep_MissingAppId(t *testing.T) {
	step := steps.NewWorkshopTakedownStep("takedown_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "publishedFileId": "1"})
	if err == nil {
		t.Fatal("expected error for missing appId")
	}
}

func TestWorkshopTakedownStep_MissingPublishedFileId(t *testing.T) {
	step := steps.NewWorkshopTakedownStep("takedown_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1"})
	if err == nil {
		t.Fatal("expected error for missing publishedFileId")
	}
}

func TestWorkshopTakedownStep_SteamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	step := steps.NewWorkshopTakedownStep("takedown_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey": "key", "appId": "1",
			"publishedFileId": "1", "baseUrl": srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error from API 403")
	}
}
