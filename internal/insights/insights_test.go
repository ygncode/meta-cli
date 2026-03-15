package insights_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/insights"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestGetPageInsights(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "111/insights") {
			t.Errorf("expected path to contain 111/insights, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("metric") != "page_impressions,page_engaged_users" {
			t.Errorf("unexpected metric param: %s", r.URL.Query().Get("metric"))
		}
		if r.URL.Query().Get("period") != "day" {
			t.Errorf("unexpected period param: %s", r.URL.Query().Get("period"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"name":   "page_impressions",
					"period": "day",
					"title":  "Daily Total Impressions",
					"id":     "111/insights/page_impressions/day",
					"values": []map[string]any{
						{"value": 100, "end_time": "2026-03-14T08:00:00+0000"},
						{"value": 150, "end_time": "2026-03-15T08:00:00+0000"},
					},
				},
				{
					"name":   "page_engaged_users",
					"period": "day",
					"title":  "Daily Engaged Users",
					"id":     "111/insights/page_engaged_users/day",
					"values": []map[string]any{
						{"value": 20, "end_time": "2026-03-14T08:00:00+0000"},
						{"value": 35, "end_time": "2026-03-15T08:00:00+0000"},
					},
				},
			},
		})
	})
	defer srv.Close()

	svc := insights.New(client)
	metrics, err := svc.GetPageInsights(context.Background(), "111", "page_impressions,page_engaged_users", "day")
	if err != nil {
		t.Fatalf("GetPageInsights: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Name != "page_impressions" {
		t.Errorf("expected page_impressions, got %s", metrics[0].Name)
	}
	if len(metrics[0].Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(metrics[0].Values))
	}
	if metrics[0].Values[0].Value != 100 {
		t.Errorf("expected value 100, got %d", metrics[0].Values[0].Value)
	}
	if metrics[1].Name != "page_engaged_users" {
		t.Errorf("expected page_engaged_users, got %s", metrics[1].Name)
	}
	if metrics[1].Values[1].Value != 35 {
		t.Errorf("expected value 35, got %d", metrics[1].Values[1].Value)
	}
}

func TestGetPageInsightsEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	})
	defer srv.Close()

	svc := insights.New(client)
	metrics, err := svc.GetPageInsights(context.Background(), "111", "page_impressions", "day")
	if err != nil {
		t.Fatalf("GetPageInsights: %v", err)
	}
	if len(metrics) != 0 {
		t.Fatalf("expected 0 metrics, got %d", len(metrics))
	}
}

func TestGetPageInsightsError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid metric", "code": 100},
		})
	})
	defer srv.Close()

	svc := insights.New(client)
	_, err := svc.GetPageInsights(context.Background(), "111", "invalid_metric", "day")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetPostInsights(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "111_222/insights") {
			t.Errorf("expected path to contain 111_222/insights, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("metric") != "post_impressions,post_clicks" {
			t.Errorf("unexpected metric param: %s", r.URL.Query().Get("metric"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"name":   "post_impressions",
					"period": "lifetime",
					"title":  "Lifetime Post Impressions",
					"id":     "111_222/insights/post_impressions/lifetime",
					"values": []map[string]any{
						{"value": 500, "end_time": "2026-03-15T08:00:00+0000"},
					},
				},
				{
					"name":   "post_clicks",
					"period": "lifetime",
					"title":  "Lifetime Post Clicks",
					"id":     "111_222/insights/post_clicks/lifetime",
					"values": []map[string]any{
						{"value": 42, "end_time": "2026-03-15T08:00:00+0000"},
					},
				},
			},
		})
	})
	defer srv.Close()

	svc := insights.New(client)
	metrics, err := svc.GetPostInsights(context.Background(), "111_222", "post_impressions,post_clicks")
	if err != nil {
		t.Fatalf("GetPostInsights: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Name != "post_impressions" {
		t.Errorf("expected post_impressions, got %s", metrics[0].Name)
	}
	if metrics[0].Values[0].Value != 500 {
		t.Errorf("expected value 500, got %d", metrics[0].Values[0].Value)
	}
	if metrics[1].Name != "post_clicks" {
		t.Errorf("expected post_clicks, got %s", metrics[1].Name)
	}
	if metrics[1].Values[0].Value != 42 {
		t.Errorf("expected value 42, got %d", metrics[1].Values[0].Value)
	}
}

func TestGetPostInsightsError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := insights.New(client)
	_, err := svc.GetPostInsights(context.Background(), "111_222", "invalid_metric")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFlatten(t *testing.T) {
	metrics := []insights.InsightMetric{
		{
			Name:   "page_impressions",
			Period: "day",
			Values: []insights.InsightValue{
				{Value: 100, EndTime: "2026-03-14T08:00:00+0000"},
				{Value: 150, EndTime: "2026-03-15T08:00:00+0000"},
			},
		},
		{
			Name:   "page_engaged_users",
			Period: "day",
			Values: []insights.InsightValue{
				{Value: 20, EndTime: "2026-03-14T08:00:00+0000"},
				{Value: 35, EndTime: "2026-03-15T08:00:00+0000"},
			},
		},
	}

	rows := insights.Flatten(metrics)
	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	if rows[0].Metric != "page_impressions" || rows[0].Value != 100 {
		t.Errorf("row 0: expected page_impressions/100, got %s/%d", rows[0].Metric, rows[0].Value)
	}
	if rows[1].Metric != "page_impressions" || rows[1].Value != 150 {
		t.Errorf("row 1: expected page_impressions/150, got %s/%d", rows[1].Metric, rows[1].Value)
	}
	if rows[2].Metric != "page_engaged_users" || rows[2].Value != 20 {
		t.Errorf("row 2: expected page_engaged_users/20, got %s/%d", rows[2].Metric, rows[2].Value)
	}
	if rows[3].Metric != "page_engaged_users" || rows[3].Value != 35 {
		t.Errorf("row 3: expected page_engaged_users/35, got %s/%d", rows[3].Metric, rows[3].Value)
	}
	if rows[0].Period != "day" {
		t.Errorf("expected period day, got %s", rows[0].Period)
	}
	if rows[0].EndTime != "2026-03-14T08:00:00+0000" {
		t.Errorf("expected end_time 2026-03-14T08:00:00+0000, got %s", rows[0].EndTime)
	}
}

func TestFlattenEmpty(t *testing.T) {
	rows := insights.Flatten(nil)
	if len(rows) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(rows))
	}
}
