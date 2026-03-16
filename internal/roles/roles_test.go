package roles_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/roles"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestRolesList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "assigned_users") {
			t.Errorf("expected path to contain assigned_users, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "user_1", "name": "Admin", "tasks": []string{"MANAGE", "CREATE_CONTENT"}},
				{"id": "user_2", "name": "Editor", "tasks": []string{"CREATE_CONTENT"}},
			},
		})
	})
	defer srv.Close()

	svc := roles.New(client)
	list, err := svc.List(context.Background(), "111")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 users, got %d", len(list))
	}
	if list[0].Name != "Admin" {
		t.Errorf("expected Admin, got %s", list[0].Name)
	}
	if list[0].Tasks != "MANAGE,CREATE_CONTENT" {
		t.Errorf("expected MANAGE,CREATE_CONTENT, got %s", list[0].Tasks)
	}
}

func TestRolesListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad request", "code": 100},
		})
	})
	defer srv.Close()

	svc := roles.New(client)
	_, err := svc.List(context.Background(), "111")
	if err == nil {
		t.Error("expected error")
	}
}

func TestRolesAssign(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "assigned_users") {
			t.Errorf("expected path to contain assigned_users, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("user") != "user_1" {
			t.Errorf("expected user=user_1, got %s", r.FormValue("user"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := roles.New(client)
	if err := svc.Assign(context.Background(), "111", "user_1", []string{"MANAGE", "CREATE_CONTENT"}); err != nil {
		t.Fatalf("Assign: %v", err)
	}
}

func TestRolesAssignFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := roles.New(client)
	err := svc.Assign(context.Background(), "111", "user_1", []string{"MANAGE"})
	if err == nil {
		t.Error("expected error when assign returns success=false")
	}
}

func TestRolesRemove(t *testing.T) {
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

	svc := roles.New(client)
	if err := svc.Remove(context.Background(), "111", "user_1"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestRolesRemoveFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := roles.New(client)
	err := svc.Remove(context.Background(), "111", "user_1")
	if err == nil {
		t.Error("expected error when remove returns success=false")
	}
}
