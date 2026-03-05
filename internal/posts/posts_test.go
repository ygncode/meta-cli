package posts_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/posts"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestPostsList(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":            "111_001",
					"message":       "Hello world",
					"created_time":  "2024-01-01T00:00:00+0000",
					"permalink_url": "https://facebook.com/111/posts/001",
					"likes":         map[string]any{"summary": map[string]any{"total_count": 5}},
					"comments":      map[string]any{"summary": map[string]any{"total_count": 3}},
					"shares":        map[string]any{"count": 1},
				},
				{
					"id":            "111_002",
					"message":       "No engagement",
					"created_time":  "2024-01-02T00:00:00+0000",
					"permalink_url": "https://facebook.com/111/posts/002",
				},
			},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	list, err := svc.List(context.Background(), "111", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(list))
	}
	if list[0].Likes != 5 {
		t.Errorf("expected 5 likes, got %d", list[0].Likes)
	}
	if list[0].Comments != 3 {
		t.Errorf("expected 3 comments, got %d", list[0].Comments)
	}
	if list[0].Shares != 1 {
		t.Errorf("expected 1 share, got %d", list[0].Shares)
	}
	// Post without engagement stats
	if list[1].Likes != 0 {
		t.Errorf("expected 0 likes for second post, got %d", list[1].Likes)
	}
}

func TestPostsCreateText(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "111_new"})
	})
	defer srv.Close()

	svc := posts.New(client)
	result, err := svc.CreateText(context.Background(), "111", "My post")
	if err != nil {
		t.Fatalf("CreateText: %v", err)
	}
	if result.ID != "111_new" {
		t.Errorf("expected 111_new, got %s", result.ID)
	}
}

func TestPostsCreateLink(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if !strings.Contains(r.Form.Get("link"), "example.com") {
			t.Errorf("expected link in body")
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "111_link"})
	})
	defer srv.Close()

	svc := posts.New(client)
	result, err := svc.CreateLink(context.Background(), "111", "check this", "https://example.com")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if result.ID != "111_link" {
		t.Errorf("expected 111_link, got %s", result.ID)
	}
}

func TestPostsCreateLinkNoMessage(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "111_link2"})
	})
	defer srv.Close()

	svc := posts.New(client)
	result, err := svc.CreateLink(context.Background(), "111", "", "https://example.com")
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if result.ID != "111_link2" {
		t.Errorf("expected 111_link2, got %s", result.ID)
	}
}

func TestPostsCreatePhoto(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart, got %s", ct)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "photo_1"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(tmpFile, []byte("fake jpg"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreatePhoto(context.Background(), "111", "caption", tmpFile)
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}
	if result.ID != "photo_1" {
		t.Errorf("expected photo_1, got %s", result.ID)
	}
}

func TestPostsDelete(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := posts.New(client)
	if err := svc.Delete(context.Background(), "111_001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestPostsDeleteFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := posts.New(client)
	err := svc.Delete(context.Background(), "111_001")
	if err == nil {
		t.Error("expected error when delete returns success=false")
	}
}

func TestPostsListError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	_, err := svc.List(context.Background(), "111", 10)
	if err == nil {
		t.Error("expected error")
	}
}
