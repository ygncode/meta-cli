package posts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	result, err := svc.CreateText(context.Background(), "111", "My post", nil)
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
	result, err := svc.CreateLink(context.Background(), "111", "check this", "https://example.com", nil)
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
	result, err := svc.CreateLink(context.Background(), "111", "", "https://example.com", nil)
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
	result, err := svc.CreatePhoto(context.Background(), "111", "caption", tmpFile, nil)
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}
	if result.ID != "photo_1" {
		t.Errorf("expected photo_1, got %s", result.ID)
	}
}

func TestPostsCreatePhotos(t *testing.T) {
	callCount := 0
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		callCount++
		ct := r.Header.Get("Content-Type")
		if strings.Contains(ct, "multipart/form-data") {
			// Unpublished photo upload
			r.ParseMultipartForm(1 << 20)
			if r.FormValue("published") != "false" {
				t.Errorf("expected published=false, got %s", r.FormValue("published"))
			}
			if r.FormValue("temporary") != "true" {
				t.Errorf("expected temporary=true, got %s", r.FormValue("temporary"))
			}
			json.NewEncoder(w).Encode(map[string]string{"id": fmt.Sprintf("media_%d", callCount)})
		} else {
			// Feed post with attached_media
			r.ParseForm()
			if r.FormValue("message") != "album" {
				t.Errorf("expected message=album, got %s", r.FormValue("message"))
			}
			if r.FormValue("attached_media[0]") == "" || r.FormValue("attached_media[1]") == "" {
				t.Errorf("expected attached_media[0] and [1]")
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "111_album"})
		}
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	files := []string{
		filepath.Join(tmpDir, "a.jpg"),
		filepath.Join(tmpDir, "b.jpg"),
	}
	for _, f := range files {
		os.WriteFile(f, []byte("fake jpg"), 0o644)
	}

	svc := posts.New(client)
	result, err := svc.CreatePhotos(context.Background(), "111", "album", files, nil)
	if err != nil {
		t.Fatalf("CreatePhotos: %v", err)
	}
	if result.ID != "111_album" {
		t.Errorf("expected 111_album, got %s", result.ID)
	}
	// 2 photo uploads + 1 feed post = 3 calls
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}

func TestPostsUpdate(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.FormValue("message") != "Updated text" {
			t.Errorf("expected message=Updated text, got %s", r.FormValue("message"))
		}
		// Verify the path contains the post ID
		if !strings.Contains(r.URL.Path, "111_001") {
			t.Errorf("expected path to contain 111_001, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := posts.New(client)
	if err := svc.Update(context.Background(), "111_001", "Updated text"); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestPostsUpdateFailed(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := posts.New(client)
	err := svc.Update(context.Background(), "111_001", "Updated text")
	if err == nil {
		t.Error("expected error when update returns success=false")
	}
}

func TestPostsUpdateAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid post", "code": 100},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	err := svc.Update(context.Background(), "111_001", "Updated text")
	if err == nil {
		t.Error("expected error on API error")
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

func TestPostsCreateTextScheduled(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("published") != "false" {
			t.Errorf("expected published=false, got %s", r.FormValue("published"))
		}
		if r.FormValue("scheduled_publish_time") == "" {
			t.Error("expected scheduled_publish_time to be set")
		}
		if r.FormValue("message") != "Scheduled post" {
			t.Errorf("expected message=Scheduled post, got %s", r.FormValue("message"))
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "111_sched"})
	})
	defer srv.Close()

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreateText(context.Background(), "111", "Scheduled post", opts)
	if err != nil {
		t.Fatalf("CreateText scheduled: %v", err)
	}
	if result.ID != "111_sched" {
		t.Errorf("expected 111_sched, got %s", result.ID)
	}
}

func TestPostsCreateLinkScheduled(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("published") != "false" {
			t.Errorf("expected published=false, got %s", r.FormValue("published"))
		}
		if r.FormValue("scheduled_publish_time") == "" {
			t.Error("expected scheduled_publish_time to be set")
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "111_link_sched"})
	})
	defer srv.Close()

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreateLink(context.Background(), "111", "check this", "https://example.com", opts)
	if err != nil {
		t.Fatalf("CreateLink scheduled: %v", err)
	}
	if result.ID != "111_link_sched" {
		t.Errorf("expected 111_link_sched, got %s", result.ID)
	}
}

func TestPostsCreatePhotoScheduled(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("published") != "false" {
			t.Errorf("expected published=false, got %s", r.FormValue("published"))
		}
		if r.FormValue("scheduled_publish_time") == "" {
			t.Error("expected scheduled_publish_time to be set")
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "photo_sched"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(tmpFile, []byte("fake jpg"), 0o644)

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreatePhoto(context.Background(), "111", "caption", tmpFile, opts)
	if err != nil {
		t.Fatalf("CreatePhoto scheduled: %v", err)
	}
	if result.ID != "photo_sched" {
		t.Errorf("expected photo_sched, got %s", result.ID)
	}
}

func TestPostsCreatePhotosScheduled(t *testing.T) {
	callCount := 0
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		ct := r.Header.Get("Content-Type")
		if strings.Contains(ct, "multipart/form-data") {
			json.NewEncoder(w).Encode(map[string]string{"id": fmt.Sprintf("media_%d", callCount)})
		} else {
			r.ParseForm()
			if r.FormValue("published") != "false" {
				t.Errorf("expected published=false on feed post, got %s", r.FormValue("published"))
			}
			if r.FormValue("scheduled_publish_time") == "" {
				t.Error("expected scheduled_publish_time on feed post")
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "111_album_sched"})
		}
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	files := []string{
		filepath.Join(tmpDir, "a.jpg"),
		filepath.Join(tmpDir, "b.jpg"),
	}
	for _, f := range files {
		os.WriteFile(f, []byte("fake jpg"), 0o644)
	}

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreatePhotos(context.Background(), "111", "album", files, opts)
	if err != nil {
		t.Fatalf("CreatePhotos scheduled: %v", err)
	}
	if result.ID != "111_album_sched" {
		t.Errorf("expected 111_album_sched, got %s", result.ID)
	}
}

func TestPostsCreateScheduledTooSoon(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach API")
	})

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(5 * time.Minute)}
	_, err := svc.CreateText(context.Background(), "111", "too soon", opts)
	if err == nil {
		t.Error("expected error for schedule time too soon")
	}
	if !strings.Contains(err.Error(), "at least 10 minutes") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostsCreateScheduledTooFar(t *testing.T) {
	_, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach API")
	})

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(76 * 24 * time.Hour)}
	_, err := svc.CreateText(context.Background(), "111", "too far", opts)
	if err == nil {
		t.Error("expected error for schedule time too far")
	}
	if !strings.Contains(err.Error(), "within 75 days") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostsListScheduled(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "scheduled_posts") {
			t.Errorf("expected path to contain scheduled_posts, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":                      "111_sched_001",
					"message":                 "Future post",
					"scheduled_publish_time":  "2026-04-01T09:00:00+0000",
					"created_time":            "2026-03-15T10:00:00+0000",
				},
			},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	list, err := svc.ListScheduled(context.Background(), "111", 10)
	if err != nil {
		t.Fatalf("ListScheduled: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 scheduled post, got %d", len(list))
	}
	if list[0].ID != "111_sched_001" {
		t.Errorf("expected 111_sched_001, got %s", list[0].ID)
	}
	if list[0].Message != "Future post" {
		t.Errorf("expected Future post, got %s", list[0].Message)
	}
}

func TestPostsListScheduledError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	_, err := svc.ListScheduled(context.Background(), "111", 10)
	if err == nil {
		t.Error("expected error")
	}
}
