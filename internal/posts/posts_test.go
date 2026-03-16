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
	result, err := svc.CreateText(context.Background(), "111", "My post", nil, nil)
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
	result, err := svc.CreateLink(context.Background(), "111", "check this", "https://example.com", nil, nil)
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
	result, err := svc.CreateLink(context.Background(), "111", "", "https://example.com", nil, nil)
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
	result, err := svc.CreatePhoto(context.Background(), "111", "caption", tmpFile, nil, nil)
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
	result, err := svc.CreateText(context.Background(), "111", "Scheduled post", opts, nil)
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
	result, err := svc.CreateLink(context.Background(), "111", "check this", "https://example.com", opts, nil)
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
	result, err := svc.CreatePhoto(context.Background(), "111", "caption", tmpFile, opts, nil)
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
	_, err := svc.CreateText(context.Background(), "111", "too soon", opts, nil)
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
	_, err := svc.CreateText(context.Background(), "111", "too far", opts, nil)
	if err == nil {
		t.Error("expected error for schedule time too far")
	}
	if !strings.Contains(err.Error(), "within 75 days") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPostsCreateVideo(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/videos") {
			t.Errorf("expected path to contain /videos, got %s", r.URL.Path)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart, got %s", ct)
		}
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("description") != "Watch this!" {
			t.Errorf("expected description=Watch this!, got %s", r.FormValue("description"))
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "video_1"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake video"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreateVideo(context.Background(), "111", posts.VideoOpts{
		FilePath: tmpFile,
		Message:  "Watch this!",
	}, nil)
	if err != nil {
		t.Fatalf("CreateVideo: %v", err)
	}
	if result.ID != "video_1" {
		t.Errorf("expected video_1, got %s", result.ID)
	}
}

func TestPostsCreateVideoWithTitle(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("title") != "My Video" {
			t.Errorf("expected title=My Video, got %s", r.FormValue("title"))
		}
		if r.FormValue("description") != "Description" {
			t.Errorf("expected description=Description, got %s", r.FormValue("description"))
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "video_2"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake video"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreateVideo(context.Background(), "111", posts.VideoOpts{
		FilePath: tmpFile,
		Title:    "My Video",
		Message:  "Description",
	}, nil)
	if err != nil {
		t.Fatalf("CreateVideo: %v", err)
	}
	if result.ID != "video_2" {
		t.Errorf("expected video_2, got %s", result.ID)
	}
}

func TestPostsCreateVideoWithThumbnail(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1 << 20)
		if _, _, err := r.FormFile("source"); err != nil {
			t.Errorf("expected source file: %v", err)
		}
		if _, _, err := r.FormFile("thumb"); err != nil {
			t.Errorf("expected thumb file: %v", err)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "video_3"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	videoFile := filepath.Join(tmpDir, "clip.mp4")
	thumbFile := filepath.Join(tmpDir, "thumb.jpg")
	os.WriteFile(videoFile, []byte("fake video"), 0o644)
	os.WriteFile(thumbFile, []byte("fake thumb"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreateVideo(context.Background(), "111", posts.VideoOpts{
		FilePath:  videoFile,
		Title:     "My Video",
		Message:   "Description",
		Thumbnail: thumbFile,
	}, nil)
	if err != nil {
		t.Fatalf("CreateVideo: %v", err)
	}
	if result.ID != "video_3" {
		t.Errorf("expected video_3, got %s", result.ID)
	}
}

func TestPostsCreateVideoScheduled(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1 << 20)
		if r.FormValue("published") != "false" {
			t.Errorf("expected published=false, got %s", r.FormValue("published"))
		}
		if r.FormValue("scheduled_publish_time") == "" {
			t.Error("expected scheduled_publish_time to be set")
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "video_sched"})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake video"), 0o644)

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreateVideo(context.Background(), "111", posts.VideoOpts{
		FilePath: tmpFile,
		Message:  "Scheduled video",
	}, opts)
	if err != nil {
		t.Fatalf("CreateVideo scheduled: %v", err)
	}
	if result.ID != "video_sched" {
		t.Errorf("expected video_sched, got %s", result.ID)
	}
}

func TestPostsCreateVideoAPIError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid video", "code": 100},
		})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake video"), 0o644)

	svc := posts.New(client)
	_, err := svc.CreateVideo(context.Background(), "111", posts.VideoOpts{
		FilePath: tmpFile,
		Message:  "test",
	}, nil)
	if err == nil {
		t.Error("expected error on API error")
	}
}

func TestPostsCreateReel(t *testing.T) {
	callCount := 0
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			// Step 1: init
			r.ParseForm()
			if r.FormValue("upload_phase") != "start" {
				t.Errorf("expected upload_phase=start, got %s", r.FormValue("upload_phase"))
			}
			if !strings.Contains(r.URL.Path, "/video_reels") {
				t.Errorf("expected path to contain /video_reels, got %s", r.URL.Path)
			}
			// Return upload_url pointing to same test server
			json.NewEncoder(w).Encode(map[string]string{
				"video_id":   "vid_123",
				"upload_url": r.URL.Query().Get("_srv") + "/upload/video",
			})
		case 2:
			// Step 2: binary upload
			if r.Header.Get("Authorization") != "OAuth test_token" {
				t.Errorf("expected OAuth header, got %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/octet-stream" {
				t.Errorf("expected octet-stream, got %s", r.Header.Get("Content-Type"))
			}
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		case 3:
			// Step 3: finish
			r.ParseForm()
			if r.FormValue("upload_phase") != "finish" {
				t.Errorf("expected upload_phase=finish, got %s", r.FormValue("upload_phase"))
			}
			if r.FormValue("video_id") != "vid_123" {
				t.Errorf("expected video_id=vid_123, got %s", r.FormValue("video_id"))
			}
			if r.FormValue("description") != "Check this out!" {
				t.Errorf("expected description=Check this out!, got %s", r.FormValue("description"))
			}
			if r.FormValue("published") != "true" {
				t.Errorf("expected published=true, got %s", r.FormValue("published"))
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "reel_001", "post_id": "111_reel_001"})
		}
	})
	defer srv.Close()

	// Patch: make init return upload_url pointing to test server
	// We'll use a custom handler that injects the srv.URL
	srv2, client2 := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	_ = srv2
	_ = client2
	// Actually, let's use a different approach: the test server itself serves all 3 steps
	// The init step needs to return an upload_url that points back to the test server.
	// But we don't know srv.URL inside the handler at definition time.
	// Solution: close this server and create a new one where we can reference the URL.
	srv.Close()

	var srvURL string
	callCount = 0
	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case callCount == 1 && strings.Contains(r.URL.Path, "/video_reels"):
			// Step 1: init
			r.ParseForm()
			if r.FormValue("upload_phase") != "start" {
				t.Errorf("expected upload_phase=start, got %s", r.FormValue("upload_phase"))
			}
			json.NewEncoder(w).Encode(map[string]string{
				"video_id":   "vid_123",
				"upload_url": srvURL + "/upload/video",
			})
		case callCount == 2 && strings.Contains(r.URL.Path, "/upload/video"):
			// Step 2: binary upload
			if r.Header.Get("Authorization") != "OAuth test_token" {
				t.Errorf("expected OAuth header, got %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/octet-stream" {
				t.Errorf("expected octet-stream, got %s", r.Header.Get("Content-Type"))
			}
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		case callCount == 3 && strings.Contains(r.URL.Path, "/video_reels"):
			// Step 3: finish
			r.ParseForm()
			if r.FormValue("upload_phase") != "finish" {
				t.Errorf("expected upload_phase=finish, got %s", r.FormValue("upload_phase"))
			}
			if r.FormValue("video_id") != "vid_123" {
				t.Errorf("expected video_id=vid_123, got %s", r.FormValue("video_id"))
			}
			if r.FormValue("description") != "Check this out!" {
				t.Errorf("expected description=Check this out!, got %s", r.FormValue("description"))
			}
			if r.FormValue("published") != "true" {
				t.Errorf("expected published=true, got %s", r.FormValue("published"))
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "reel_001", "post_id": "111_reel_001"})
		default:
			t.Errorf("unexpected call %d to %s", callCount, r.URL.Path)
		}
	}))
	defer srv3.Close()
	srvURL = srv3.URL

	client = graph.NewWithHTTPClient(srv3.URL, "test_token", srv3.Client())

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake reel video"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreateReel(context.Background(), "111", posts.ReelOpts{
		FilePath: tmpFile,
		Message:  "Check this out!",
	}, nil)
	if err != nil {
		t.Fatalf("CreateReel: %v", err)
	}
	if result.ID != "reel_001" {
		t.Errorf("expected reel_001, got %s", result.ID)
	}
	if result.PostID != "111_reel_001" {
		t.Errorf("expected 111_reel_001, got %s", result.PostID)
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}

func TestPostsCreateReelWithTitle(t *testing.T) {
	var srvURL string
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case callCount == 1:
			json.NewEncoder(w).Encode(map[string]string{
				"video_id":   "vid_456",
				"upload_url": srvURL + "/upload/video",
			})
		case callCount == 2:
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		case callCount == 3:
			r.ParseForm()
			if r.FormValue("title") != "My Reel" {
				t.Errorf("expected title=My Reel, got %s", r.FormValue("title"))
			}
			if r.FormValue("description") != "Description" {
				t.Errorf("expected description=Description, got %s", r.FormValue("description"))
			}
			json.NewEncoder(w).Encode(map[string]string{"id": "reel_002"})
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake reel"), 0o644)

	svc := posts.New(client)
	result, err := svc.CreateReel(context.Background(), "111", posts.ReelOpts{
		FilePath: tmpFile,
		Title:    "My Reel",
		Message:  "Description",
	}, nil)
	if err != nil {
		t.Fatalf("CreateReel with title: %v", err)
	}
	if result.ID != "reel_002" {
		t.Errorf("expected reel_002, got %s", result.ID)
	}
}

func TestPostsCreateReelScheduled(t *testing.T) {
	var srvURL string
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case callCount == 1:
			json.NewEncoder(w).Encode(map[string]string{
				"video_id":   "vid_789",
				"upload_url": srvURL + "/upload/video",
			})
		case callCount == 2:
			json.NewEncoder(w).Encode(map[string]bool{"success": true})
		case callCount == 3:
			r.ParseForm()
			if r.FormValue("published") != "false" {
				t.Errorf("expected published=false, got %s", r.FormValue("published"))
			}
			if r.FormValue("scheduled_publish_time") == "" {
				t.Error("expected scheduled_publish_time to be set")
			}
			// Should NOT have published=true
			json.NewEncoder(w).Encode(map[string]string{"id": "reel_sched"})
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake reel"), 0o644)

	svc := posts.New(client)
	opts := &posts.ScheduleOpts{PublishTime: time.Now().Add(1 * time.Hour)}
	result, err := svc.CreateReel(context.Background(), "111", posts.ReelOpts{
		FilePath: tmpFile,
		Message:  "Coming soon!",
	}, opts)
	if err != nil {
		t.Fatalf("CreateReel scheduled: %v", err)
	}
	if result.ID != "reel_sched" {
		t.Errorf("expected reel_sched, got %s", result.ID)
	}
}

func TestPostsCreateReelInitError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "invalid page", "code": 100},
		})
	})
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake reel"), 0o644)

	svc := posts.New(client)
	_, err := svc.CreateReel(context.Background(), "111", posts.ReelOpts{
		FilePath: tmpFile,
		Message:  "test",
	}, nil)
	if err == nil {
		t.Error("expected error on init failure")
	}
	if !strings.Contains(err.Error(), "init reel upload") {
		t.Errorf("expected init error context, got: %v", err)
	}
}

func TestPostsCreateReelUploadError(t *testing.T) {
	var srvURL string
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case callCount == 1:
			json.NewEncoder(w).Encode(map[string]string{
				"video_id":   "vid_err",
				"upload_url": srvURL + "/upload/video",
			})
		case callCount == 2:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "upload failed", "code": 500},
			})
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	client := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake reel"), 0o644)

	svc := posts.New(client)
	_, err := svc.CreateReel(context.Background(), "111", posts.ReelOpts{
		FilePath: tmpFile,
		Message:  "test",
	}, nil)
	if err == nil {
		t.Error("expected error on upload failure")
	}
	if !strings.Contains(err.Error(), "upload reel video") {
		t.Errorf("expected upload error context, got: %v", err)
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

func TestPostsListVisitor(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "visitor_posts") {
			t.Errorf("expected path to contain visitor_posts, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "visitor_001",
					"message":      "Great page!",
					"from":         map[string]any{"name": "Alice"},
					"created_time": "2026-01-01T00:00:00+0000",
				},
			},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	list, err := svc.ListVisitor(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("ListVisitor: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 visitor post, got %d", len(list))
	}
	if list[0].From != "Alice" {
		t.Errorf("expected Alice, got %s", list[0].From)
	}
}

func TestPostsListVisitorError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	_, err := svc.ListVisitor(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
	}
}

func TestPostsListTagged(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "tagged") {
			t.Errorf("expected path to contain tagged, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":           "tagged_001",
					"message":      "Check out @page",
					"from":         map[string]any{"name": "Bob"},
					"created_time": "2026-01-02T00:00:00+0000",
				},
			},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	list, err := svc.ListTagged(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("ListTagged: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 tagged post, got %d", len(list))
	}
	if list[0].From != "Bob" {
		t.Errorf("expected Bob, got %s", list[0].From)
	}
}

func TestPostsListTaggedError(t *testing.T) {
	srv, client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := posts.New(client)
	_, err := svc.ListTagged(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
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
