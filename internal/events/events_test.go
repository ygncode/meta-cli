package events_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/events"
	"github.com/ygncode/meta-cli/internal/graph"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestEventsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "events") {
			t.Errorf("expected path to contain events, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":          "event_001",
					"name":        "Launch Party",
					"description": "Product launch",
					"start_time":  "2026-04-01T18:00:00+0000",
					"end_time":    "2026-04-01T22:00:00+0000",
					"place":       map[string]any{"name": "HQ"},
				},
			},
		})
	})
	defer srv.Close()

	svc := events.New(client)
	list, err := svc.List(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 event, got %d", len(list))
	}
	if list[0].ID != "event_001" {
		t.Errorf("expected event_001, got %s", list[0].ID)
	}
	if list[0].Name != "Launch Party" {
		t.Errorf("expected Launch Party, got %s", list[0].Name)
	}
	if list[0].Place != "HQ" {
		t.Errorf("expected HQ, got %s", list[0].Place)
	}
}

func TestEventsListEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	defer srv.Close()

	svc := events.New(client)
	list, err := svc.List(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 events, got %d", len(list))
	}
}

func TestEventsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := events.New(client)
	_, err := svc.List(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
	}
}
