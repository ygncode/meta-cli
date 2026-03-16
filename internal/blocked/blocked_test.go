package blocked_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/blocked"
	"github.com/ygncode/meta-cli/internal/graph"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestBlockedList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "blocked") {
			t.Errorf("expected path to contain blocked, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "user_1", "name": "Spammer"},
				{"id": "user_2", "name": "Troll"},
			},
		})
	})
	defer srv.Close()

	svc := blocked.New(client)
	list, err := svc.List(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 blocked users, got %d", len(list))
	}
	if list[0].Name != "Spammer" {
		t.Errorf("expected Spammer, got %s", list[0].Name)
	}
}

func TestBlockedListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := blocked.New(client)
	_, err := svc.List(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
	}
}

func TestBlockedBlock(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "blocked") {
			t.Errorf("expected path to contain blocked, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("user") != "user_1" {
			t.Errorf("expected user=user_1, got %s", r.FormValue("user"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := blocked.New(client)
	if err := svc.Block(context.Background(), "111", "user_1"); err != nil {
		t.Fatalf("Block: %v", err)
	}
}

func TestBlockedBlockFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := blocked.New(client)
	err := svc.Block(context.Background(), "111", "user_1")
	if err == nil {
		t.Error("expected error when block returns success=false")
	}
}

func TestBlockedUnblock(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Query().Get("user") != "user_1" {
			t.Errorf("expected user query param=user_1, got %s", r.URL.Query().Get("user"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := blocked.New(client)
	if err := svc.Unblock(context.Background(), "111", "user_1"); err != nil {
		t.Fatalf("Unblock: %v", err)
	}
}

func TestBlockedUnblockFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := blocked.New(client)
	err := svc.Unblock(context.Background(), "111", "user_1")
	if err == nil {
		t.Error("expected error when unblock returns success=false")
	}
}
