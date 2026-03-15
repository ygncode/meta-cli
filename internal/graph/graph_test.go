package graph_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
)

func TestNew(t *testing.T) {
	c := graph.New("v21.0", "test_token")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithToken(t *testing.T) {
	c := graph.New("v21.0", "original")
	c2 := c.WithToken("new_token")
	if c2 == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGraphErrorError(t *testing.T) {
	e := &graph.GraphError{Message: "bad request", Type: "OAuthException", Code: 100}
	s := e.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIErrorError(t *testing.T) {
	e := &graph.APIError{StatusCode: 400, Graph: &graph.GraphError{Message: "bad", Code: 100}}
	s := e.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}

	e2 := &graph.APIError{StatusCode: 500}
	s2 := e2.Error()
	if s2 == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	ge := &graph.GraphError{Message: "test", Code: 190}
	e := &graph.APIError{StatusCode: 401, Graph: ge}
	unwrapped := e.Unwrap()
	if unwrapped != ge {
		t.Errorf("expected unwrapped to be GraphError")
	}

	e2 := &graph.APIError{StatusCode: 500}
	if e2.Unwrap() != nil {
		t.Error("expected nil unwrap when no Graph error")
	}
}

func TestIsTokenExpired(t *testing.T) {
	expired := &graph.APIError{
		StatusCode: 401,
		Graph:      &graph.GraphError{Code: 190, Message: "token expired"},
	}
	if !graph.IsTokenExpired(expired) {
		t.Error("expected IsTokenExpired to be true for code 190")
	}

	notExpired := &graph.APIError{
		StatusCode: 400,
		Graph:      &graph.GraphError{Code: 100, Message: "bad request"},
	}
	if graph.IsTokenExpired(notExpired) {
		t.Error("expected IsTokenExpired to be false for code 100")
	}

	if graph.IsTokenExpired(errors.New("random error")) {
		t.Error("expected IsTokenExpired to be false for non-APIError")
	}
}

func TestIsPermissionDenied(t *testing.T) {
	denied := &graph.APIError{
		StatusCode: 403,
		Graph:      &graph.GraphError{Code: 200, Message: "permission denied"},
	}
	if !graph.IsPermissionDenied(denied) {
		t.Error("expected IsPermissionDenied to be true for code 200")
	}

	notDenied := &graph.APIError{
		StatusCode: 400,
		Graph:      &graph.GraphError{Code: 100},
	}
	if graph.IsPermissionDenied(notDenied) {
		t.Error("expected IsPermissionDenied to be false for code 100")
	}
}

// newTestClient creates a graph.Client pointing at the test server.
func newTestClient(t *testing.T, srv *httptest.Server) *graph.Client {
	t.Helper()
	return graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
}

func TestClientGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		tok := r.URL.Query().Get("access_token")
		if tok != "test_token" {
			t.Errorf("expected test_token, got %s", tok)
		}
		fields := r.URL.Query().Get("fields")
		if fields != "id,name" {
			t.Errorf("expected fields=id,name, got %s", fields)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Test"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.Get(context.Background(), "me", url.Values{"fields": {"id,name"}}, &out)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out["id"] != "123" {
		t.Errorf("expected id=123, got %s", out["id"])
	}
}

func TestClientGetNilParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	if err := c.Get(context.Background(), "me", nil, &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestClientGetAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"message": "Invalid token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.Get(context.Background(), "me", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *graph.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Graph == nil || apiErr.Graph.Code != 190 {
		t.Errorf("expected graph error code 190, got %+v", apiErr.Graph)
	}
}

func TestClientGetNonGraphError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.Get(context.Background(), "me", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *graph.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Graph != nil {
		t.Errorf("expected nil Graph error for non-JSON response")
	}
}

func TestClientPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Errorf("expected form content type, got %s", ct)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "post_123"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.Post(context.Background(), "page/feed", url.Values{"message": {"hello"}}, &out)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if out["id"] != "post_123" {
		t.Errorf("expected post_123, got %s", out["id"])
	}
}

func TestClientDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]bool
	err := c.Delete(context.Background(), "123_456", &out)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !out["success"] {
		t.Error("expected success=true")
	}
}

func TestClientPostMultipart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart content type, got %s", ct)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "photo_123"})
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	os.WriteFile(tmpFile, []byte("fake image data"), 0o644)

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.PostMultipart(context.Background(), "page/photos",
		map[string]string{"message": "look at this"},
		tmpFile, &out)
	if err != nil {
		t.Fatalf("PostMultipart: %v", err)
	}
	if out["id"] != "photo_123" {
		t.Errorf("expected photo_123, got %s", out["id"])
	}
}

func TestClientPostMultipartMissingFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "x"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.PostMultipart(context.Background(), "page/photos", nil, "/no/such/file.jpg", &out)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestClientPostMultipartFiles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart content type, got %s", ct)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if r.FormValue("description") != "my video" {
			t.Errorf("expected description=my video, got %s", r.FormValue("description"))
		}
		if _, _, err := r.FormFile("source"); err != nil {
			t.Errorf("expected source file: %v", err)
		}
		if _, _, err := r.FormFile("thumb"); err != nil {
			t.Errorf("expected thumb file: %v", err)
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "video_123"})
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	videoFile := filepath.Join(tmpDir, "clip.mp4")
	thumbFile := filepath.Join(tmpDir, "thumb.jpg")
	os.WriteFile(videoFile, []byte("fake video data"), 0o644)
	os.WriteFile(thumbFile, []byte("fake thumb data"), 0o644)

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.PostMultipartFiles(context.Background(), "page/videos",
		map[string]string{"description": "my video"},
		map[string]string{"source": videoFile, "thumb": thumbFile},
		&out)
	if err != nil {
		t.Fatalf("PostMultipartFiles: %v", err)
	}
	if out["id"] != "video_123" {
		t.Errorf("expected video_123, got %s", out["id"])
	}
}

func TestClientPostMultipartFilesMissingFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "x"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	var out map[string]string
	err := c.PostMultipartFiles(context.Background(), "page/videos",
		nil,
		map[string]string{"source": "/no/such/file.mp4"},
		&out)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestClientPostBinary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		auth := r.Header.Get("Authorization")
		if auth != "OAuth test_token" {
			t.Errorf("expected Authorization: OAuth test_token, got %s", auth)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/octet-stream" {
			t.Errorf("expected application/octet-stream, got %s", ct)
		}
		offset := r.Header.Get("offset")
		if offset != "0" {
			t.Errorf("expected offset=0, got %s", offset)
		}
		fileSize := r.Header.Get("file_size")
		if fileSize != "16" {
			t.Errorf("expected file_size=16, got %s", fileSize)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "fake video bytes" {
			t.Errorf("expected body=fake video bytes, got %s", string(body))
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clip.mp4")
	os.WriteFile(tmpFile, []byte("fake video bytes"), 0o644)

	c := newTestClient(t, srv)
	var out map[string]bool
	err := c.PostBinary(context.Background(), srv.URL+"/upload/video", tmpFile, &out)
	if err != nil {
		t.Fatalf("PostBinary: %v", err)
	}
	if !out["success"] {
		t.Error("expected success=true")
	}
}

func TestClientPostBinaryMissingFile(t *testing.T) {
	c := graph.New("v21.0", "test_token")
	var out map[string]bool
	err := c.PostBinary(context.Background(), "https://example.com/upload", "/no/such/file.mp4", &out)
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestClientGetCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := graph.New("v21.0", "test_token")
	var out map[string]string
	err := c.Get(ctx, "me", nil, &out)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestClientGetNilOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	// nil output should not panic
	err := c.Get(context.Background(), "me", nil, nil)
	if err != nil {
		t.Fatalf("Get with nil out: %v", err)
	}
}
