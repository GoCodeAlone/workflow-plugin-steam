package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSteamAuth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ISteamUserAuth/AuthenticateUserTicket/v1/" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"params": map[string]any{
					"result":          "OK",
					"steamid":         "76561198000000001",
					"ownersteamid":    "76561198000000001",
					"vacbanned":       false,
					"publisherbanned": false,
				},
			},
		})
	}))
	defer srv.Close()

	step := newSteamAuthStep("test_auth")
	result, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "testapikey",
			"appId":   "12345",
			"ticket":  "aabbccddeeff",
			"baseUrl": srv.URL,
		}, nil, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output["steamId"] != "76561198000000001" {
		t.Errorf("steamId = %v, want 76561198000000001", result.Output["steamId"])
	}
	if result.Output["result"] != "OK" {
		t.Errorf("result = %v, want OK", result.Output["result"])
	}
	if result.Output["vacBanned"] != false {
		t.Errorf("vacBanned = %v, want false", result.Output["vacBanned"])
	}
}

func TestSteamAuth_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": map[string]any{
				"error": map[string]any{
					"errorcode": 102,
					"errordesc": "invalid ticket",
				},
			},
		})
	}))
	defer srv.Close()

	step := newSteamAuthStep("test_auth")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "testapikey",
			"appId":   "12345",
			"ticket":  "badbadticket",
			"baseUrl": srv.URL,
		}, nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSteamAuth_MissingFields(t *testing.T) {
	step := newSteamAuthStep("test_auth")

	_, err := step.Execute(context.Background(), nil, nil, nil, nil, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing apiKey")
	}

	_, err = step.Execute(context.Background(), nil, nil, nil, nil,
		map[string]any{"apiKey": "k", "appId": "1"})
	if err == nil {
		t.Fatal("expected error for missing ticket")
	}
}

func TestSteamAuth_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	step := newSteamAuthStep("test_auth")
	_, err := step.Execute(context.Background(), nil, nil,
		map[string]any{
			"apiKey":  "k",
			"appId":   "1",
			"ticket":  "t",
			"baseUrl": srv.URL,
		}, nil, nil)
	if err == nil {
		t.Fatal("expected error on HTTP 403")
	}
}
