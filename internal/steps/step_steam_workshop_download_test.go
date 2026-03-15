package steps_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-steam/internal/steps"
)

// buildTestWorkshopZip creates a minimal valid .steamworkshop zip.
func buildTestWorkshopZip() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	manifest := `{"schemaVersion":1,"itemType":"ruleset","title":"DL Test","description":"","tags":[],"previewImagePath":"","gameTypes":["gwent"],"minPlayers":2,"maxPlayers":2,"author":"76561198012345678","version":"1.0.0"}`
	w, _ := zw.Create("manifest.json")
	w.Write([]byte(manifest))
	zw.Close()
	return buf.Bytes()
}

// makeDownloadServer creates a fake Steam API + CDN server that returns file details and serves a zip.
func makeDownloadServer(t *testing.T, zipData []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/ISteamRemoteStorage/GetPublishedFileDetails/v1/" ||
			r.URL.Path == "/ISteamRemoteStorage/SubscribePublishedFile/v1/" {
			json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"result": 1,
					"publishedfiledetails": []any{
						map[string]any{
							"publishedfileid": "99999",
							"file_url":        "REPLACE_WITH_SELF/download/file.steamworkshop",
							"time_updated":    float64(1700000000),
						},
					},
				},
			})
			return
		}
		if r.URL.Path == "/download/file.steamworkshop" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(zipData)
			return
		}
		http.NotFound(w, r)
	}))
}

func TestWorkshopDownloadStep_Success(t *testing.T) {
	zipData := buildTestWorkshopZip()
	srv := makeDownloadServer(t, zipData)
	defer srv.Close()

	// We need to patch the file_url to point to our test server.
	// Use a custom server that returns the test server's URL in file_url.
	var srvURL string
	patchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/ISteamRemoteStorage/SubscribePublishedFile/v1/" {
			json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
			return
		}
		if r.URL.Path == "/ISteamRemoteStorage/GetPublishedFileDetails/v1/" {
			json.NewEncoder(w).Encode(map[string]any{
				"response": map[string]any{
					"result": 1,
					"publishedfiledetails": []any{
						map[string]any{
							"publishedfileid": "99999",
							"file_url":        srvURL + "/download/file.steamworkshop",
							"time_updated":    float64(1700000000),
						},
					},
				},
			})
			return
		}
		if r.URL.Path == "/download/file.steamworkshop" {
			w.Header().Set("Content-Type", "application/zip")
			w.Write(zipData)
			return
		}
		http.NotFound(w, r)
	}))
	defer patchSrv.Close()
	srvURL = patchSrv.URL

	installDir := t.TempDir()
	step := steps.NewWorkshopDownloadStep("dl_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "76561198012345678",
			"publishedFileId": "99999",
			"installDir":      installDir,
			"validateRuleset": false,
			"baseUrl":         patchSrv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["installed"] != true {
		t.Errorf("installed = %v, want true", result.Output["installed"])
	}
	if result.Output["itemDir"] == "" {
		t.Error("itemDir should not be empty")
	}
}

func TestWorkshopDownloadStep_SubscribeOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"result": 1}})
	}))
	defer srv.Close()

	step := steps.NewWorkshopDownloadStep("dl_test")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":          "key",
			"appId":           "12345",
			"steamId":         "s",
			"publishedFileId": "99999",
			"subscribeOnly":   true,
			"baseUrl":         srv.URL,
		}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["installed"] != false {
		t.Errorf("installed should be false when subscribeOnly=true")
	}
}

func TestWorkshopDownloadStep_MissingFileId(t *testing.T) {
	step := steps.NewWorkshopDownloadStep("dl_test")
	_, err := step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1", "steamId": "s"})
	if err == nil {
		t.Fatal("expected error for missing publishedFileId")
	}
}

// Ensure io is used
var _ = io.EOF
