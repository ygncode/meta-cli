package hooks_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ygncode/meta-cli/internal/debounce"
	"github.com/ygncode/meta-cli/internal/hooks"
)

func TestCallAgentSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "runId": "run_123"})
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "test_token", srv.Client())
	err := c.CallAgent(context.Background(), "hello", "user_1")
	if err != nil {
		t.Fatalf("CallAgent: %v", err)
	}
}

func TestCallAgentRequestFormat(t *testing.T) {
	var gotBody map[string]any
	var gotHeaders http.Header

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "runId": "run_123"})
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "my_token", srv.Client())
	c.CallAgent(context.Background(), "test prompt", "psid_42")

	if gotBody["message"] != "test prompt" {
		t.Errorf("expected message 'test prompt', got %v", gotBody["message"])
	}
	if gotBody["name"] != "FB Messenger" {
		t.Errorf("expected name 'FB Messenger', got %v", gotBody["name"])
	}
	if gotBody["deliver"] != false {
		t.Errorf("expected deliver false, got %v", gotBody["deliver"])
	}
	if gotBody["sessionKey"] != "hook:fb:psid_42" {
		t.Errorf("expected sessionKey 'hook:fb:psid_42', got %v", gotBody["sessionKey"])
	}
	if ct := gotHeaders.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestCallAgentAuthHeader(t *testing.T) {
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "runId": "run_123"})
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "secret_token", srv.Client())
	c.CallAgent(context.Background(), "hello", "user_1")

	if gotAuth != "Bearer secret_token" {
		t.Errorf("expected 'Bearer secret_token', got %s", gotAuth)
	}
}

func TestCallAgentServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "token", srv.Client())
	err := c.CallAgent(context.Background(), "hello", "user_1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention 500, got: %s", err.Error())
	}
}

func TestCallAgentUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "bad_token", srv.Client())
	err := c.CallAgent(context.Background(), "hello", "user_1")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention 401, got: %s", err.Error())
	}
}

func TestCallAgentTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := hooks.NewClientWithHTTP(srv.URL, "token", srv.Client())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := c.CallAgent(ctx, "hello", "user_1")
	if err == nil {
		t.Fatal("expected error for timeout")
	}
}

func TestCallAgentBadEndpoint(t *testing.T) {
	c := hooks.NewClient("http://invalid.invalid:99999/hooks/agent", "token")
	err := c.CallAgent(context.Background(), "hello", "user_1")
	if err == nil {
		t.Fatal("expected error for bad endpoint")
	}
}

func TestRenderPrompt(t *testing.T) {
	tmpl := "User {{.PSID}} on {{.PageID}} says:{{range .Messages}} [{{.Text}}]{{end}}"
	msgs := []debounce.Message{
		{ID: "1", Text: "hello"},
		{ID: "2", Text: "world"},
	}

	result, err := hooks.RenderPrompt(tmpl, "user_42", "page_99", msgs)
	if err != nil {
		t.Fatalf("RenderPrompt: %v", err)
	}

	if !strings.Contains(result, "user_42") {
		t.Errorf("expected PSID in output, got: %s", result)
	}
	if !strings.Contains(result, "page_99") {
		t.Errorf("expected PageID in output, got: %s", result)
	}
	if !strings.Contains(result, "[hello]") {
		t.Errorf("expected message text in output, got: %s", result)
	}
	if !strings.Contains(result, "[world]") {
		t.Errorf("expected second message text in output, got: %s", result)
	}
}

func TestRenderPromptDefault(t *testing.T) {
	msgs := []debounce.Message{
		{ID: "1", Text: "How do I reset my password?"},
	}

	result, err := hooks.RenderPrompt("", "user_1", "page_1", msgs)
	if err != nil {
		t.Fatalf("RenderPrompt: %v", err)
	}

	if !strings.Contains(result, "PSID: user_1") {
		t.Errorf("expected PSID in default template output, got: %s", result)
	}
	if !strings.Contains(result, "page_1") {
		t.Errorf("expected PageID in default template output, got: %s", result)
	}
	if !strings.Contains(result, "How do I reset my password?") {
		t.Errorf("expected message text in default template output, got: %s", result)
	}
	if !strings.Contains(result, "meta-cli skill") {
		t.Errorf("expected skill instruction in default template output, got: %s", result)
	}
}

func TestRenderPromptInvalid(t *testing.T) {
	_, err := hooks.RenderPrompt("{{.Invalid", "user_1", "page_1", nil)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}
