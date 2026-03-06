package messenger_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ygncode/meta-cli/internal/graph"
	"github.com/ygncode/meta-cli/internal/messenger"
)

func newTestGraphClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *graph.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := graph.NewWithHTTPClient(srv.URL, "test_token", srv.Client())
	return srv, c
}

func TestServiceSend(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_resp",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	mid, err := svc.Send(context.Background(), "user_1", "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if mid != "mid_resp" {
		t.Errorf("expected message ID mid_resp, got %s", mid)
	}
}

func TestServiceSendError(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	_, err := svc.Send(context.Background(), "user_1", "hello")
	if err == nil {
		t.Error("expected error")
	}
}

func TestServiceSendCancelledContext(t *testing.T) {
	client := graph.New("v21.0", "test_token")
	svc := messenger.NewService(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Send(ctx, "user_1", "hello")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestServiceSubscribeWebhook(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	if err := svc.SubscribeWebhook(context.Background()); err != nil {
		t.Fatalf("SubscribeWebhook: %v", err)
	}
}

func TestServiceSubscribeWebhookFailure(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	if err := svc.SubscribeWebhook(context.Background()); err == nil {
		t.Error("expected error for success=false")
	}
}

func TestServiceSendTyping(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct{}{})
	})
	defer srv.Close()

	svc := messenger.NewService(client)

	if err := svc.SendTyping(context.Background(), "user_1", true); err != nil {
		t.Fatalf("SendTyping(on): %v", err)
	}
	if err := svc.SendTyping(context.Background(), "user_1", false); err != nil {
		t.Fatalf("SendTyping(off): %v", err)
	}
}

func TestServiceSendTypingCancelledContext(t *testing.T) {
	client := graph.New("v21.0", "test_token")
	svc := messenger.NewService(client)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.SendTyping(ctx, "user_1", true)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
