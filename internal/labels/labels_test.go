package labels_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/labels"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestLabelsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "custom_labels") {
			t.Errorf("expected path to contain custom_labels, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "label_001", "name": "VIP"},
				{"id": "label_002", "name": "Support"},
				{"id": "label_003", "name": "Spam"},
			},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	list, err := svc.List(context.Background(), "111")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(list))
	}
	if list[0].ID != "label_001" {
		t.Errorf("expected label_001, got %s", list[0].ID)
	}
	if list[0].Name != "VIP" {
		t.Errorf("expected VIP, got %s", list[0].Name)
	}
	if list[1].Name != "Support" {
		t.Errorf("expected Support, got %s", list[1].Name)
	}
	if list[2].Name != "Spam" {
		t.Errorf("expected Spam, got %s", list[2].Name)
	}
}

func TestLabelsListEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	list, err := svc.List(context.Background(), "111")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 labels, got %d", len(list))
	}
}

func TestLabelsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	_, err := svc.List(context.Background(), "111")
	if err == nil {
		t.Error("expected error")
	}
}

func TestLabelsCreate(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "custom_labels") {
			t.Errorf("expected path to contain custom_labels, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("name") != "VIP" {
			t.Errorf("expected name=VIP, got %s", r.FormValue("name"))
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "label_new"})
	})
	defer srv.Close()

	svc := labels.New(client)
	id, err := svc.Create(context.Background(), "111", "VIP")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != "label_new" {
		t.Errorf("expected label_new, got %s", id)
	}
}

func TestLabelsCreateError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	_, err := svc.Create(context.Background(), "111", "VIP")
	if err == nil {
		t.Error("expected error")
	}
}

func TestLabelsDelete(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "label_001") {
			t.Errorf("expected path to contain label_001, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := labels.New(client)
	if err := svc.Delete(context.Background(), "label_001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestLabelsDeleteFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Delete(context.Background(), "label_001")
	if err == nil {
		t.Error("expected error when delete returns success=false")
	}
}

func TestLabelsDeleteAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "label not found", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Delete(context.Background(), "label_001")
	if err == nil {
		t.Error("expected error on API error")
	}
}

func TestLabelsAssign(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "label_001/label") {
			t.Errorf("expected path to contain label_001/label, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("user") != "psid_123" {
			t.Errorf("expected user=psid_123, got %s", r.FormValue("user"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := labels.New(client)
	if err := svc.Assign(context.Background(), "label_001", "psid_123"); err != nil {
		t.Fatalf("Assign: %v", err)
	}
}

func TestLabelsAssignFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Assign(context.Background(), "label_001", "psid_123")
	if err == nil {
		t.Error("expected error when assign returns success=false")
	}
}

func TestLabelsAssignAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid user", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Assign(context.Background(), "label_001", "psid_123")
	if err == nil {
		t.Error("expected error on API error")
	}
}

func TestLabelsRemove(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "label_001/label") {
			t.Errorf("expected path to contain label_001/label, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("user") != "psid_123" {
			t.Errorf("expected user query param=psid_123, got %s", r.URL.Query().Get("user"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := labels.New(client)
	if err := svc.Remove(context.Background(), "label_001", "psid_123"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestLabelsRemoveFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Remove(context.Background(), "label_001", "psid_123")
	if err == nil {
		t.Error("expected error when remove returns success=false")
	}
}

func TestLabelsRemoveAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	err := svc.Remove(context.Background(), "label_001", "psid_123")
	if err == nil {
		t.Error("expected error on API error")
	}
}

func TestLabelsListByUser(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "psid_123/custom_labels") {
			t.Errorf("expected path to contain psid_123/custom_labels, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "label_001", "name": "VIP"},
			},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	list, err := svc.ListByUser(context.Background(), "psid_123")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 label, got %d", len(list))
	}
	if list[0].Name != "VIP" {
		t.Errorf("expected VIP, got %s", list[0].Name)
	}
}

func TestLabelsListByUserEmpty(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	list, err := svc.ListByUser(context.Background(), "psid_123")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 labels, got %d", len(list))
	}
}

func TestLabelsListByUserError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid", "code": 100},
		})
	})
	defer srv.Close()

	svc := labels.New(client)
	_, err := svc.ListByUser(context.Background(), "psid_123")
	if err == nil {
		t.Error("expected error")
	}
}
