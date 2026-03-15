package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

func TestCollectionQueryStep_ReturnsItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"collectiondetails": []any{
					map[string]any{
						"publishedfileid": "col1",
						"title":           "My Collection",
						"file_description": "A bundle of rulesets",
						"children": []any{
							map[string]any{"publishedfileid": "item1"},
							map[string]any{"publishedfileid": "item2"},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	step := steps.NewWorkshopCollectionQueryStep("cq_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":       "key",
			"appId":        "12345",
			"collectionId": "col1",
			"baseUrl":      srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, ok := result.Output["items"].([]any)
	if !ok {
		t.Fatalf("items not in output, got %v", result.Output)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	if result.Output["itemCount"] != 2 {
		t.Errorf("itemCount = %v, want 2", result.Output["itemCount"])
	}
}

func TestCollectionQueryStep_EmptyCollection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"collectiondetails": []any{
					map[string]any{
						"publishedfileid": "col1",
						"title":           "Empty",
						"children":        []any{},
					},
				},
			},
		})
	}))
	defer srv.Close()

	step := steps.NewWorkshopCollectionQueryStep("cq_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{"apiKey": "key", "appId": "12345", "collectionId": "col1", "baseUrl": srv.URL}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["itemCount"] != 0 {
		t.Errorf("itemCount = %v, want 0", result.Output["itemCount"])
	}
}

func TestCollectionQueryStep_MissingCollectionId(t *testing.T) {
	step := steps.NewWorkshopCollectionQueryStep("cq_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1"})
	if err == nil {
		t.Fatal("expected error for missing collectionId")
	}
}

func TestCollectionQueryStep_NotACollection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not a collection", http.StatusBadRequest)
	}))
	defer srv.Close()

	step := steps.NewWorkshopCollectionQueryStep("cq_test")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{"apiKey": "key", "appId": "1", "collectionId": "bad", "baseUrl": srv.URL}, nil, nil)
	if err == nil {
		t.Fatal("expected error when API returns error")
	}
}
