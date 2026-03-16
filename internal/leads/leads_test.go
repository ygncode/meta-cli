package leads_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/leads"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestLeadsCreateForm(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "leadgen_forms") {
			t.Errorf("expected path to contain leadgen_forms, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "form_001"})
	})
	defer srv.Close()

	svc := leads.New(client)
	payload := json.RawMessage(`{"name":"Test Form","questions":[{"type":"FULL_NAME"}]}`)
	id, err := svc.CreateForm(context.Background(), "111", payload)
	if err != nil {
		t.Fatalf("CreateForm: %v", err)
	}
	if id != "form_001" {
		t.Errorf("expected form_001, got %s", id)
	}
}

func TestLeadsCreateFormError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := leads.New(client)
	payload := json.RawMessage(`{"name":"Test Form"}`)
	_, err := svc.CreateForm(context.Background(), "111", payload)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLeadsCreateFormInvalidJSON(t *testing.T) {
	client := graph.New("v21.0", "test_token")
	svc := leads.New(client)
	_, err := svc.CreateForm(context.Background(), "111", json.RawMessage(`invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLeadsListLeads(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "leads") {
			t.Errorf("expected path to contain leads, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "lead_001",
					"created_time": "2026-01-01T00:00:00+0000",
					"field_data": []map[string]any{
						{"name": "full_name", "values": []string{"John Doe"}},
						{"name": "email", "values": []string{"john@example.com"}},
					},
				},
			},
		})
	})
	defer srv.Close()

	svc := leads.New(client)
	list, err := svc.ListLeads(context.Background(), "form_001", 50)
	if err != nil {
		t.Fatalf("ListLeads: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 lead, got %d", len(list))
	}
	if list[0].ID != "lead_001" {
		t.Errorf("expected lead_001, got %s", list[0].ID)
	}
}

func TestLeadsListLeadsEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	})
	defer srv.Close()

	svc := leads.New(client)
	list, err := svc.ListLeads(context.Background(), "form_001", 50)
	if err != nil {
		t.Fatalf("ListLeads: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 leads, got %d", len(list))
	}
}

func TestLeadsListLeadsError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := leads.New(client)
	_, err := svc.ListLeads(context.Background(), "form_001", 50)
	if err == nil {
		t.Error("expected error")
	}
}
