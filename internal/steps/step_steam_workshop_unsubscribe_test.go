package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestWorkshopUnsubscribeStep_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	dir := t.TempDir()

	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"publishedFileId": "111",
			"installDir":      dir,
			"removeFiles":     true,
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["unsubscribed"] != true {
		t.Errorf("unsubscribed = %v, want true", result.Output["unsubscribed"])
	}
	if result.Output["filesRemoved"] != true {
		t.Errorf("filesRemoved = %v, want true", result.Output["filesRemoved"])
	}
}

func TestWorkshopUnsubscribeStep_NoRemoveFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	dir := t.TempDir()
	itemDir := filepath.Join(dir, "222")
	if err := os.MkdirAll(itemDir, 0o755); err != nil {
		t.Fatal(err)
	}

	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"publishedFileId": "222",
			"installDir":      dir,
			"removeFiles":     false,
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["unsubscribed"] != true {
		t.Errorf("unsubscribed = %v, want true", result.Output["unsubscribed"])
	}
	if result.Output["filesRemoved"] != false {
		t.Errorf("filesRemoved = %v, want false", result.Output["filesRemoved"])
	}
	// item dir should still exist
	if _, err := os.Stat(itemDir); os.IsNotExist(err) {
		t.Error("item dir was removed but removeFiles=false")
	}
}

func TestWorkshopUnsubscribeStep_MissingApiKey(t *testing.T) {
	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"appId": "1", "steamId": "s", "publishedFileId": "1"})
	if err == nil {
		t.Fatal("expected error for missing apiKey")
	}
}

func TestWorkshopUnsubscribeStep_MissingAppId(t *testing.T) {
	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "steamId": "s", "publishedFileId": "1"})
	if err == nil {
		t.Fatal("expected error for missing appId")
	}
}

func TestWorkshopUnsubscribeStep_MissingSteamId(t *testing.T) {
	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "publishedFileId": "1"})
	if err == nil {
		t.Fatal("expected error for missing steamId")
	}
}

func TestWorkshopUnsubscribeStep_MissingPublishedFileId(t *testing.T) {
	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s"})
	if err == nil {
		t.Fatal("expected error for missing publishedFileId")
	}
}

func TestWorkshopUnsubscribeStep_SteamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	step := steps.NewWorkshopUnsubscribeStep("unsub_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey": "key", "appId": "1", "steamId": "s",
			"publishedFileId": "1", "baseUrl": srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error from API 403")
	}
}
