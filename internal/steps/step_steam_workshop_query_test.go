package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func makeQueryServer(t *testing.T, response map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func workshopQueryResult(items []any, total int) map[string]any {
	return map[string]any{
		"response": map[string]any{
			"total": float64(total),
			"publishedfiledetails": items,
		},
	}
}

func sampleItem(id string) map[string]any {
	return map[string]any{
		"publishedfileid":  id,
		"title":            "Test Item " + id,
		"file_description": "Desc",
		"tags":             []any{map[string]any{"tag": "gwent"}},
		"preview_url":      "https://example.com/preview.png",
		"subscriptions":    float64(100),
		"vote_data": map[string]any{
			"score":      float64(0.85),
			"votes_up":   float64(85),
			"votes_down": float64(15),
		},
		"time_created": float64(1700000000),
		"time_updated": float64(1700100000),
		"creator":      "76561198012345678",
	}
}

func TestWorkshopQueryStep_SearchByTag(t *testing.T) {
	srv := makeQueryServer(t, workshopQueryResult([]any{sampleItem("111")}, 1))
	defer srv.Close()

	step := steps.NewWorkshopQueryStep("q_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":    "key",
			"appId":     "12345",
			"queryType": "ranked_by_vote",
			"tags":      "gwent",
			"baseUrl":   srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, ok := result.Output["items"].([]any)
	if !ok || len(items) == 0 {
		t.Errorf("expected items in result, got %v", result.Output["items"])
	}
	if result.Output["totalCount"] != 1 {
		t.Errorf("totalCount = %v, want 1", result.Output["totalCount"])
	}
}

func TestWorkshopQueryStep_SearchByText(t *testing.T) {
	srv := makeQueryServer(t, workshopQueryResult([]any{sampleItem("222"), sampleItem("333")}, 2))
	defer srv.Close()

	step := steps.NewWorkshopQueryStep("q_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":     "key",
			"appId":      "12345",
			"queryType":  "text_search",
			"searchText": "turbo gwent",
			"baseUrl":    srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := result.Output["items"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestWorkshopQueryStep_PaginationParams(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workshopQueryResult([]any{}, 0))
	}))
	defer srv.Close()

	step := steps.NewWorkshopQueryStep("q_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":   "key",
			"appId":    "12345",
			"page":     float64(3),
			"pageSize": float64(10),
			"baseUrl":  srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedQuery == "" {
		t.Error("no query captured")
	}
}

func TestWorkshopQueryStep_GetByFileId(t *testing.T) {
	srv := makeQueryServer(t, workshopQueryResult([]any{sampleItem("456")}, 1))
	defer srv.Close()

	step := steps.NewWorkshopQueryStep("q_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"queryType":       "get_by_id",
			"publishedFileId": "456",
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := result.Output["items"].([]any)
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestWorkshopQueryStep_EmptyResult(t *testing.T) {
	srv := makeQueryServer(t, workshopQueryResult([]any{}, 0))
	defer srv.Close()

	step := steps.NewWorkshopQueryStep("q_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{"apiKey": "key", "appId": "12345", "baseUrl": srv.URL}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["totalCount"] != 0 {
		t.Errorf("totalCount = %v, want 0", result.Output["totalCount"])
	}
}

func TestWorkshopQueryStep_MissingApiKey(t *testing.T) {
	step := steps.NewWorkshopQueryStep("q_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"appId": "12345"})
	if err == nil {
		t.Fatal("expected error for missing apiKey")
	}
}
