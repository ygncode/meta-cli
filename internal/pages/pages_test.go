package pages_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/pages"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestPagesList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{
				{"id": "111", "name": "Page One"},
				{"id": "222", "name": "Page Two"},
			},
		})
	})
	defer srv.Close()

	svc := pages.New(client)
	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(list))
	}
	if list[0].ID != "111" || list[0].Name != "Page One" {
		t.Errorf("unexpected first page: %+v", list[0])
	}
}

func TestPagesListEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	defer srv.Close()

	svc := pages.New(client)
	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 pages, got %d", len(list))
	}
}

func TestPagesListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad token", "code": 190},
		})
	})
	defer srv.Close()

	svc := pages.New(client)
	_, err := svc.List(context.Background())
	if err == nil {
		t.Error("expected error")
	}
}

func TestPagesInfo(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":                  "111",
			"name":                "Test Page",
			"about":               "About text",
			"description":         "Description text",
			"category":            "Software",
			"phone":               "+1234567890",
			"website":             "https://example.com",
			"emails":              []string{"test@example.com"},
			"fan_count":           1000,
			"followers_count":     950,
			"verification_status": "not_verified",
		})
	})
	defer srv.Close()

	svc := pages.New(client)
	info, err := svc.Info(context.Background(), "111")
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.ID != "111" {
		t.Errorf("expected 111, got %s", info.ID)
	}
	if info.Name != "Test Page" {
		t.Errorf("expected Test Page, got %s", info.Name)
	}
	if info.FanCount != 1000 {
		t.Errorf("expected 1000, got %d", info.FanCount)
	}
	if info.FollowersCount != 950 {
		t.Errorf("expected 950, got %d", info.FollowersCount)
	}
}

func TestPagesInfoError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := pages.New(client)
	_, err := svc.Info(context.Background(), "111")
	if err == nil {
		t.Error("expected error")
	}
}
