package ratings_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/ratings"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestRatingsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "ratings") {
			t.Errorf("expected path to contain ratings, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"reviewer":     map[string]any{"name": "Alice"},
					"rating":       5,
					"review_text":  "Excellent!",
					"created_time": "2026-01-01T00:00:00+0000",
				},
				{
					"reviewer":     map[string]any{"name": "Bob"},
					"rating":       4,
					"review_text":  "Good",
					"created_time": "2026-01-02T00:00:00+0000",
				},
			},
		})
	})
	defer srv.Close()

	svc := ratings.New(client)
	list, err := svc.List(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 ratings, got %d", len(list))
	}
	if list[0].ReviewerName != "Alice" {
		t.Errorf("expected Alice, got %s", list[0].ReviewerName)
	}
	if list[0].Rating != 5 {
		t.Errorf("expected 5, got %d", list[0].Rating)
	}
}

func TestRatingsListEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	defer srv.Close()

	svc := ratings.New(client)
	list, err := svc.List(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 ratings, got %d", len(list))
	}
}

func TestRatingsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := ratings.New(client)
	_, err := svc.List(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
	}
}

func TestRatingsSummary(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"overall_star_rating": 4.5,
			"rating_count":       120,
		})
	})
	defer srv.Close()

	svc := ratings.New(client)
	summary, err := svc.Summary(context.Background(), "111")
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.StarRating != 4.5 {
		t.Errorf("expected 4.5, got %f", summary.StarRating)
	}
	if summary.RatingCount != 120 {
		t.Errorf("expected 120, got %d", summary.RatingCount)
	}
}

func TestRatingsSummaryError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := ratings.New(client)
	_, err := svc.Summary(context.Background(), "111")
	if err == nil {
		t.Error("expected error")
	}
}
