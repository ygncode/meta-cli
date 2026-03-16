package reactions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/reactions"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestReactionsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "reactions") {
			t.Errorf("expected path to contain reactions, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "user_1", "name": "Alice", "type": "LIKE"},
				{"id": "user_2", "name": "Bob", "type": "LOVE"},
				{"id": "user_3", "name": "Charlie", "type": "HAHA"},
			},
		})
	})
	defer srv.Close()

	svc := reactions.New(client)
	list, err := svc.List(context.Background(), "post_001", 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 reactions, got %d", len(list))
	}
	if list[0].Type != "LIKE" {
		t.Errorf("expected LIKE, got %s", list[0].Type)
	}
	if list[1].Name != "Bob" {
		t.Errorf("expected Bob, got %s", list[1].Name)
	}
}

func TestReactionsListEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	defer srv.Close()

	svc := reactions.New(client)
	list, err := svc.List(context.Background(), "post_001", 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 reactions, got %d", len(list))
	}
}

func TestReactionsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := reactions.New(client)
	_, err := svc.List(context.Background(), "post_001", 50)
	if err == nil {
		t.Error("expected error")
	}
}
