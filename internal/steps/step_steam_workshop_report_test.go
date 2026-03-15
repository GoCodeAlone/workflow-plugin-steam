package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestWorkshopReportStep_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	step := steps.NewWorkshopReportStep("rep_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"publishedFileId": "111",
			"reportType":      "spam",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["reported"] != true {
		t.Errorf("reported = %v, want true", result.Output["reported"])
	}
}

func TestWorkshopReportStep_MissingPublishedFileId(t *testing.T) {
	step := steps.NewWorkshopReportStep("rep_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s", "reportType": "spam"})
	if err == nil {
		t.Fatal("expected error for missing publishedFileId")
	}
}

func TestWorkshopReportStep_InvalidReportType(t *testing.T) {
	step := steps.NewWorkshopReportStep("rep_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s", "publishedFileId": "1", "reportType": "invalid_type"})
	if err == nil {
		t.Fatal("expected error for invalid reportType")
	}
}

func TestWorkshopReportStep_SteamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	step := steps.NewWorkshopReportStep("rep_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey": "key", "appId": "1", "steamId": "s",
			"publishedFileId": "1", "reportType": "spam", "baseUrl": srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error from API 403")
	}
}
