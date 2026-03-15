package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestWorkshopPublishStep_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"publishedfileid": "123456789",
				"result":          1,
			},
		})
	}))
	defer srv.Close()

	step := steps.NewWorkshopPublishStep("pub_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":      "test-api-key",
			"appId":       "12345",
			"steamId":     "76561198012345678",
			"packagePath": "/tmp/test.steamworkshop",
			"title":       "Turbo Gwent",
			"description": "Fast variant",
			"baseUrl":     srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["publishedFileId"] != "123456789" {
		t.Errorf("publishedFileId = %v, want 123456789", result.Output["publishedFileId"])
	}
	if result.Output["created"] != true {
		t.Errorf("created = %v, want true", result.Output["created"])
	}
	if result.Output["itemUrl"] == "" {
		t.Error("itemUrl should not be empty")
	}
}

func TestWorkshopPublishStep_MissingApiKey(t *testing.T) {
	step := steps.NewWorkshopPublishStep("pub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"appId": "12345", "steamId": "s", "packagePath": "/tmp/x", "title": "T"})
	if err == nil {
		t.Fatal("expected error for missing apiKey")
	}
}

func TestWorkshopPublishStep_MissingPackagePath(t *testing.T) {
	step := steps.NewWorkshopPublishStep("pub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "12345", "steamId": "s", "title": "T"})
	if err == nil {
		t.Fatal("expected error for missing packagePath")
	}
}

func TestWorkshopPublishStep_SteamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	step := steps.NewWorkshopPublishStep("pub_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":      "key",
			"appId":       "12345",
			"steamId":     "76561198012345678",
			"packagePath": "/tmp/test.steamworkshop",
			"title":       "Test",
			"baseUrl":     srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error from Steam API 400")
	}
}

func TestWorkshopPublishStep_UpdateExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"result": 1,
			},
		})
	}))
	defer srv.Close()

	step := steps.NewWorkshopPublishStep("pub_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"packagePath":     "/tmp/test.steamworkshop",
			"title":           "Update",
			"publishedFileId": "987654321",
			"changelog":       "v2 release",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["created"] != false {
		t.Errorf("created should be false for update, got %v", result.Output["created"])
	}
	if result.Output["publishedFileId"] != "987654321" {
		t.Errorf("publishedFileId = %v, want 987654321", result.Output["publishedFileId"])
	}
}
