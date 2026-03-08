package comments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ygncode/meta-cli/internal/comments"
	"github.com/ygncode/meta-cli/internal/graph"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestCommentsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "c_001",
					"message":      "Nice post!",
					"from":         map[string]string{"name": "Alice"},
					"created_time": "2024-01-01T00:00:00+0000",
					"like_count":   2,
					"is_hidden":    false,
				},
				{
					"id":           "c_002",
					"message":      "Spam",
					"created_time": "2024-01-02T00:00:00+0000",
					"like_count":   0,
					"is_hidden":    true,
				},
			},
		})
	})
	defer srv.Close()

	svc := comments.New(client)
	list, err := svc.List(context.Background(), "111_001", 25)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(list))
	}
	if list[0].From != "Alice" {
		t.Errorf("expected From=Alice, got %s", list[0].From)
	}
	if list[0].LikeCount != 2 {
		t.Errorf("expected 2 likes, got %d", list[0].LikeCount)
	}
	// Comment without from field
	if list[1].From != "" {
		t.Errorf("expected empty From for comment without from, got %s", list[1].From)
	}
	if !list[1].IsHidden {
		t.Error("expected second comment to be hidden")
	}
}

func TestCommentsReply(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "reply_001"})
	})
	defer srv.Close()

	svc := comments.New(client)
	id, err := svc.Reply(context.Background(), "c_001", "Thanks!")
	if err != nil {
		t.Fatalf("Reply: %v", err)
	}
	if id != "reply_001" {
		t.Errorf("expected reply_001, got %s", id)
	}
}

func TestCommentsSetHidden(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := comments.New(client)

	if err := svc.SetHidden(context.Background(), "c_001", true); err != nil {
		t.Fatalf("SetHidden(true): %v", err)
	}
	if err := svc.SetHidden(context.Background(), "c_001", false); err != nil {
		t.Fatalf("SetHidden(false): %v", err)
	}
}

func TestCommentsSetHiddenFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := comments.New(client)
	err := svc.SetHidden(context.Background(), "c_001", true)
	if err == nil {
		t.Error("expected error when hide returns success=false")
	}

	err = svc.SetHidden(context.Background(), "c_001", false)
	if err == nil {
		t.Error("expected error when unhide returns success=false")
	}
}

func TestCommentsUpdate(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.FormValue("message") != "Updated comment" {
			t.Errorf("expected message=Updated comment, got %s", r.FormValue("message"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := comments.New(client)
	if err := svc.Update(context.Background(), "c_001", "Updated comment"); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestCommentsUpdateFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := comments.New(client)
	err := svc.Update(context.Background(), "c_001", "Updated comment")
	if err == nil {
		t.Error("expected error when update returns success=false")
	}
}

func TestCommentsUpdateAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid comment", "code": 100},
		})
	})
	defer srv.Close()

	svc := comments.New(client)
	err := svc.Update(context.Background(), "c_001", "Updated comment")
	if err == nil {
		t.Error("expected error on API error")
	}
}

func TestCommentsDelete(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := comments.New(client)
	if err := svc.Delete(context.Background(), "c_001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestCommentsDeleteFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := comments.New(client)
	err := svc.Delete(context.Background(), "c_001")
	if err == nil {
		t.Error("expected error when delete returns success=false")
	}
}

func TestCommentsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := comments.New(client)
	_, err := svc.List(context.Background(), "111_001", 25)
	if err == nil {
		t.Error("expected error")
	}
}
