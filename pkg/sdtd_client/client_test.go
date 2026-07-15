/*
Copyright © 2024-2025 Tom Helander thomas.helander@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package sdtdclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"
)

func TestNewSDTDClient(t *testing.T) {
	t.Run("returns error for empty host", func(t *testing.T) {
		_, err := NewSDTDClient("", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != ErrNoHostSet {
			t.Errorf("expected ErrNoHostSet, got %v", err)
		}
	})

	t.Run("returns error for host without scheme", func(t *testing.T) {
		_, err := NewSDTDClient("example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != ErrInvalidHostScheme {
			t.Errorf("expected ErrInvalidHostScheme, got %v", err)
		}
	})

	t.Run("returns error for nil auth", func(t *testing.T) {
		_, err := NewSDTDClient("http://example.com", nil, true, slog.Default())
		if err != ErrNilAuth {
			t.Errorf("expected ErrNilAuth, got %v", err)
		}
	})

	t.Run("returns client for valid inputs", func(t *testing.T) {
		client, err := NewSDTDClient("http://example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.Host != "http://example.com" {
			t.Errorf("expected host http://example.com, got %s", client.Host)
		}
		if client.Auth == nil {
			t.Error("expected non-nil Auth")
		}
		if client.client == nil {
			t.Error("expected non-nil HTTP client")
		}
		if client.logger == nil {
			t.Error("expected non-nil logger")
		}
	})
}

func TestGetHeaders(t *testing.T) {
	client, err := NewSDTDClient("http://example.com", &SDTDAuth{
		TokenName:   "testToken",
		TokenSecret: "testSecret",
	}, true, slog.Default())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headers := client.GetHeaders()

	// http.Header is case-insensitive for Get(), but the values are set with uppercase keys
	if headers.Get("Accept") != "application/json" {
		t.Errorf("expected Accept=application/json, got %s", headers.Get("Accept"))
	}
	// Check using the exact key name as set in GetHeaders()
	if len(headers["X-SDTD-API-TOKENNAME"]) == 0 {
		t.Error("expected X-SDTD-API-TOKENNAME header to be set")
	} else if headers["X-SDTD-API-TOKENNAME"][0] != "testToken" {
		t.Errorf("expected X-SDTD-API-TOKENNAME=testToken, got %s", headers["X-SDTD-API-TOKENNAME"][0])
	}
	if len(headers["X-SDTD-API-SECRET"]) == 0 {
		t.Error("expected X-SDTD-API-SECRET header to be set")
	} else if headers["X-SDTD-API-SECRET"][0] != "testSecret" {
		t.Errorf("expected X-SDTD-API-SECRET=testSecret, got %s", headers["X-SDTD-API-SECRET"][0])
	}
}

func TestDo(t *testing.T) {
	t.Run("returns error for invalid path", func(t *testing.T) {
		client, err := NewSDTDClient("http://example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = client.Do("GET", "/api/invalid", nil, nil)
		if err == nil {
			t.Error("expected error for invalid path")
		}
	})

	t.Run("handles successful GET request", func(t *testing.T) {
		var callCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"test": "value"})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = client.Do("GET", "/api/test", nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if callCount == 0 {
			t.Error("expected server to be called")
		}
	})

	t.Run("handles successful POST request with data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			// Check that POST body contains the expected data
			if !bytes.Contains(body, []byte("test")) || !bytes.Contains(body, []byte("data")) {
				t.Errorf("expected POST body to contain test data, got %s", string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = client.Do("POST", "/api/test", nil, []byte(`{"test": "data"}`))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("handles successful DELETE request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = client.Do("DELETE", "/api/test", nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("handles non-2XX response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = client.Do("GET", "/api/test", nil, nil)
		if err != ErrNon2XXResponse {
			t.Errorf("expected ErrNon2XXResponse, got %v", err)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("unmarshals JSON response into ServerInfoResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": []map[string]any{{"name": "Test Server", "type": "string", "value": "test"}},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := ServerInfoResponse{}
		err = Get(client, "/api/test", &resp, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Errorf("expected 1 data item, got %d", len(resp.Data))
		}
	})

	t.Run("returns error on JSON unmarshal failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json{}`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := ServerInfoResponse{}
		err = Get(client, "/api/test", &resp, nil)
		if err == nil {
			t.Error("expected error on invalid JSON")
		}
	})
}

func TestPost(t *testing.T) {
	t.Run("unmarshals JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"}})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := BaseResponse{}
		err = Post(client, "/api/test", &resp, nil, []byte(`{"name": "test"}`))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("unmarshals JSON response when body is present", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"meta": {"serverTime": "2024-01-01T00:00:00Z"}}`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := BaseResponse{}
		err = Delete(client, "/api/test", &resp, nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// BaseResponse should have meta populated
		if resp.Meta.ServerTime != "2024-01-01T00:00:00Z" {
			t.Errorf("expected serverTime to be 2024-01-01T00:00:00Z, got %s", resp.Meta.ServerTime)
		}
	})

	t.Run("returns nil on empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`null`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := BaseResponse{}
		err = Delete(client, "/api/test", &resp, nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetM(t *testing.T) {
	t.Run("returns error when allocs not enabled", func(t *testing.T) {
		client, err := NewSDTDClient("http://example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp := PlayersResponseM{}
		err = GetM(client, "/api/test", &resp, nil)
		if err != ErrAllocsModNotInstalled {
			t.Errorf("expected ErrAllocsModNotInstalled, got %v", err)
		}
	})
}

func TestConnect(t *testing.T) {
	t.Run("detects Alloc's Server Fixes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First call is /api/serverinfo, second is /api/getstats
			if r.URL.Path == "/api/getstats" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
					"data": map[string]any{
						"gameTime": map[string]int{"days": 0, "hours": 0, "minutes": 0},
						"players":  0,
						"hostiles": 0,
						"animals":  0,
					},
				})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": []map[string]any{{"name": "test", "type": "string", "value": "test"}},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.Connect()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !client.allocsEnabled {
			t.Error("expected allocsEnabled to be true after Connect")
		}
		if !client.allocsChecked {
			t.Error("expected allocsChecked to be true after Connect")
		}
	})

	t.Run("connects without allocs when API fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.Connect()
		if err == nil {
			t.Error("expected error when server returns non-2XX")
		}
		if client.allocsEnabled {
			t.Error("expected allocsEnabled to remain false")
		}
	})
}

func TestGetServerInfo(t *testing.T) {
	t.Run("returns server info", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": []map[string]any{{"name": "Test Server", "type": "string", "value": "test"}},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := client.GetServerInfo()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(info.Data) != 1 {
			t.Errorf("expected 1 data item, got %d", len(info.Data))
		}
	})
}

func TestGetServerStats(t *testing.T) {
	t.Run("returns server stats", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": map[string]any{
					"gameTime": map[string]int{"days": 10, "hours": 5, "minutes": 30},
					"players":  5,
					"hostiles": 20,
					"animals":  100,
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stats, err := client.GetServerStats()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if stats.Data.GameTime.Days != 10 {
			t.Errorf("expected days=10, got %d", stats.Data.GameTime.Days)
		}
		if stats.Data.Players != 5 {
			t.Errorf("expected players=5, got %d", stats.Data.Players)
		}
	})
}

func TestGetGamePrefs(t *testing.T) {
	t.Run("returns game preferences", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": []map[string]any{
					{"name": "allowFlying", "type": "bool", "default": true, "value": false},
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		prefs, err := client.GetGamePrefs()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(prefs.Data) != 1 {
			t.Errorf("expected 1 preference, got %d", len(prefs.Data))
		}
	})
}

func TestGetUserStatus(t *testing.T) {
	t.Run("returns user status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": map[string]any{
					"username":        "testuser",
					"loggedIn":        true,
					"permissionLevel": 1,
					"permissions": []map[string]any{
						{"module": "build", "allowed": map[string]any{"GET": true}},
					},
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		status, err := client.GetUserStatus()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if status.Data.Username != "testuser" {
			t.Errorf("expected username=testuser, got %s", status.Data.Username)
		}
		if !status.Data.LoggedIn {
			t.Error("expected loggedIn=true")
		}
	})
}

func TestGetOnlinePlayers(t *testing.T) {
	t.Run("returns online players", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": map[string]any{
					"players": []map[string]any{
						{"entityId": 1, "name": "Player1", "online": true},
						{"entityId": 2, "name": "Player2", "online": false},
					},
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		players, err := client.GetOnlinePlayers()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(players.Data.Players) != 2 {
			t.Errorf("expected 2 players, got %d", len(players.Data.Players))
		}
	})
}

func TestGetLog(t *testing.T) {
	t.Run("returns log entries", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check query params
			if r.URL.Query().Get("count") != "50" {
				t.Errorf("expected count=50, got %s", r.URL.Query().Get("count"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": map[string]any{
					"entries": []map[string]any{
						{"id": 1, "msg": "Test log message", "type": "Log", "trace": "", "isotime": "2024-01-01T00:00:00Z", "uptime": "0"},
					},
					"firstLine": 1,
					"lastLine":  2,
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logs, err := client.GetLog(&[]int{50}[0], nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(logs.Data.Entries) != 1 {
			t.Errorf("expected 1 log entry, got %d", len(logs.Data.Entries))
		}
	})

	t.Run("handles nil count and firstLine", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should not have query params
			if r.URL.RawQuery != "" {
				t.Errorf("expected no query params, got %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"meta": map[string]any{"serverTime": "2024-01-01T00:00:00Z"},
				"data": map[string]any{
					"entries":   []map[string]any{},
					"firstLine": 0,
					"lastLine":  1,
				},
			})
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		logs, err := client.GetLog(nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(logs.Data.Entries) != 0 {
			t.Errorf("expected 0 log entries, got %d", len(logs.Data.Entries))
		}
	})
}

func TestAddWhitelistUser(t *testing.T) {
	t.Run("adds user to whitelist", func(t *testing.T) {
		var receivedID, receivedName string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/whitelist/user/123" {
				t.Errorf("expected path /api/whitelist/user/123, got %s", r.URL.Path)
			}
			if r.Method != "POST" {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			receivedID = strings.Split(r.URL.Path, "/")[4]
			var data map[string]string
			json.Unmarshal(body, &data)
			receivedName = data["name"]
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta": {"serverTime": "2024-01-01T00:00:00Z"}}`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.AddWhitelistUser("123", "TestUser")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if receivedID != "123" {
			t.Errorf("expected id=123, got %s", receivedID)
		}
		if receivedName != "TestUser" {
			t.Errorf("expected name=TestUser, got %s", receivedName)
		}
	})

	t.Run("returns error on JSON marshal failure", func(t *testing.T) {
		client, err := NewSDTDClient("http://example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.AddWhitelistUser("123", "")
		if err == nil {
			t.Error("expected error on empty name")
		}
	})
}

func TestDeleteWhitelistUser(t *testing.T) {
	t.Run("removes user from whitelist", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/whitelist/user/456" {
				t.Errorf("expected path /api/whitelist/user/456, got %s", r.URL.Path)
			}
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"meta": {"serverTime": "2024-01-01T00:00:00Z"}}`))
		}))
		defer server.Close()

		client, err := NewSDTDClient(server.URL, &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.DeleteWhitelistUser("456")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGetWhitelist(t *testing.T) {
	t.Run("returns nil (stub implementation)", func(t *testing.T) {
		client, err := NewSDTDClient("http://example.com", &SDTDAuth{
			TokenName:   "test",
			TokenSecret: "secret",
		}, true, slog.Default())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = client.GetWhitelist()
		if err != nil {
			t.Errorf("expected no error from stub implementation, got %v", err)
		}
	})
}
