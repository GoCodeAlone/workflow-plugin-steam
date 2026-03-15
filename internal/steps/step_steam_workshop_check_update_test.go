package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
	"github.com/GoCodeAlone/workflow-plugin-steam/internal/workshop"
)

func makeFileDetailsServer(t *testing.T, items []map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"publishedfiledetails": items,
			},
		})
	}))
}

func TestCheckUpdateStep_NoUpdates(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	srv := makeFileDetailsServer(t, []map[string]any{
		{"publishedfileid": "111", "time_updated": float64(ts.Unix())},
	})
	defer srv.Close()

	installDir := t.TempDir()
	db := workshop.NewVersionDB(filepath.Join(installDir, ".versions.json"))
	db.Set(workshop.VersionRecord{
		PublishedFileId: "111",
		LastUpdatedAt:   ts,
		InstalledAt:     ts,
		Version:         "1.0.0",
		ItemDir:         filepath.Join(installDir, "111"),
	})

	step := steps.NewWorkshopCheckUpdateStep("cu_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"installDir":      installDir,
			"publishedFileIds": []any{"111"},
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updates, _ := result.Output["updates"].([]any)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update entry, got %d", len(updates))
	}
	u := updates[0].(map[string]any)
	if u["hasUpdate"] != false {
		t.Errorf("hasUpdate should be false when timestamps match")
	}
}

func TestCheckUpdateStep_DetectsUpdate(t *testing.T) {
	tsOld := time.Unix(1700000000, 0).UTC()
	tsNew := time.Unix(1700100000, 0).UTC()
	srv := makeFileDetailsServer(t, []map[string]any{
		{"publishedfileid": "222", "time_updated": float64(tsNew.Unix())},
	})
	defer srv.Close()

	installDir := t.TempDir()
	db := workshop.NewVersionDB(filepath.Join(installDir, ".versions.json"))
	db.Set(workshop.VersionRecord{
		PublishedFileId: "222",
		LastUpdatedAt:   tsOld,
		InstalledAt:     tsOld,
		Version:         "1.0.0",
		ItemDir:         filepath.Join(installDir, "222"),
	})

	step := steps.NewWorkshopCheckUpdateStep("cu_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":     "key",
			"appId":      "12345",
			"installDir": installDir,
			"baseUrl":    srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updates, _ := result.Output["updates"].([]any)
	if len(updates) == 0 {
		t.Fatal("expected update entries")
	}
	u := updates[0].(map[string]any)
	if u["hasUpdate"] != true {
		t.Errorf("hasUpdate should be true when API timestamp is newer")
	}
}

func TestCheckUpdateStep_NewItem(t *testing.T) {
	// Item not in versiondb — treated as needs install (hasUpdate: true)
	ts := time.Unix(1700000000, 0).UTC()
	srv := makeFileDetailsServer(t, []map[string]any{
		{"publishedfileid": "333", "time_updated": float64(ts.Unix())},
	})
	defer srv.Close()

	installDir := t.TempDir()
	// No versiondb record for "333"

	step := steps.NewWorkshopCheckUpdateStep("cu_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"installDir":      installDir,
			"publishedFileIds": []any{"333"},
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updates, _ := result.Output["updates"].([]any)
	if len(updates) == 0 {
		t.Fatal("expected update entries for new item")
	}
	u := updates[0].(map[string]any)
	if u["hasUpdate"] != true {
		t.Errorf("hasUpdate should be true for item not in versiondb")
	}
}

func TestCheckUpdateStep_EmptyList(t *testing.T) {
	// No publishedFileIds + empty versiondb = empty updates
	srv := makeFileDetailsServer(t, []map[string]any{})
	defer srv.Close()

	installDir := t.TempDir()

	step := steps.NewWorkshopCheckUpdateStep("cu_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":     "key",
			"appId":      "12345",
			"installDir": installDir,
			"baseUrl":    srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updates, _ := result.Output["updates"].([]any)
	if len(updates) != 0 {
		t.Errorf("expected empty updates, got %d", len(updates))
	}
}
