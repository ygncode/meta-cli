package messenger_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestServiceSendAttachmentURL(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("message")
		if !strings.Contains(msg, "attachment") {
			t.Errorf("expected attachment in message, got %s", msg)
		}
		if !strings.Contains(msg, "https://example.com/image.jpg") {
			t.Errorf("expected URL in message payload")
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_attach_url",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	mid, err := svc.SendAttachmentURL(context.Background(), "user_1", "image", "https://example.com/image.jpg")
	if err != nil {
		t.Fatalf("SendAttachmentURL: %v", err)
	}
	if mid != "mid_attach_url" {
		t.Errorf("expected mid_attach_url, got %s", mid)
	}
}

func TestServiceSendAttachmentURLError(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	_, err := svc.SendAttachmentURL(context.Background(), "user_1", "image", "https://example.com/image.jpg")
	if err == nil {
		t.Error("expected error")
	}
}

func TestServiceSendTagged(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("messaging_type") != "MESSAGE_TAG" {
			t.Errorf("expected MESSAGE_TAG, got %s", r.FormValue("messaging_type"))
		}
		if r.FormValue("tag") != "HUMAN_AGENT" {
			t.Errorf("expected HUMAN_AGENT, got %s", r.FormValue("tag"))
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_tagged",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	mid, err := svc.SendTagged(context.Background(), "user_1", "hello", "HUMAN_AGENT")
	if err != nil {
		t.Fatalf("SendTagged: %v", err)
	}
	if mid != "mid_tagged" {
		t.Errorf("expected mid_tagged, got %s", mid)
	}
}

func TestServiceSendTaggedInvalidTag(t *testing.T) {
	client := graph.New("v21.0", "test_token")
	svc := messenger.NewService(client)
	_, err := svc.SendTagged(context.Background(), "user_1", "hello", "INVALID_TAG")
	if err == nil {
		t.Error("expected error for invalid tag")
	}
}

func TestServiceSendPrivateReply(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		recipient := r.FormValue("recipient")
		if !strings.Contains(recipient, "comment_id") {
			t.Errorf("expected comment_id in recipient, got %s", recipient)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_private",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	mid, err := svc.SendPrivateReply(context.Background(), "comment_123", "Private reply")
	if err != nil {
		t.Fatalf("SendPrivateReply: %v", err)
	}
	if mid != "mid_private" {
		t.Errorf("expected mid_private, got %s", mid)
	}
}

func TestServiceSendPrivateReplyError(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	_, err := svc.SendPrivateReply(context.Background(), "comment_123", "text")
	if err == nil {
		t.Error("expected error")
	}
}

func TestServiceSendTemplate(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("message")
		if !strings.Contains(msg, "template") {
			t.Errorf("expected template in message, got %s", msg)
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_template",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	payload := json.RawMessage(`{"template_type":"button","text":"Hello","buttons":[]}`)
	mid, err := svc.SendTemplate(context.Background(), "user_1", payload)
	if err != nil {
		t.Fatalf("SendTemplate: %v", err)
	}
	if mid != "mid_template" {
		t.Errorf("expected mid_template, got %s", mid)
	}
}

func TestServiceSendWithQuickReplies(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		msg := r.FormValue("message")
		if !strings.Contains(msg, "quick_replies") {
			t.Errorf("expected quick_replies in message, got %s", msg)
		}
		if !strings.Contains(msg, "Yes") || !strings.Contains(msg, "No") {
			t.Errorf("expected Yes and No in quick replies")
		}
		json.NewEncoder(w).Encode(map[string]string{
			"recipient_id": "user_1",
			"message_id":   "mid_qr",
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	mid, err := svc.SendWithQuickReplies(context.Background(), "user_1", "Pick one", []string{"Yes", "No"})
	if err != nil {
		t.Fatalf("SendWithQuickReplies: %v", err)
	}
	if mid != "mid_qr" {
		t.Errorf("expected mid_qr, got %s", mid)
	}
}

func TestServiceGetProfile(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"greeting": []map[string]any{{"locale": "default", "text": "Welcome!"}}},
			},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	result, err := svc.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestServiceSetGreeting(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	if err := svc.SetGreeting(context.Background(), "Welcome!"); err != nil {
		t.Fatalf("SetGreeting: %v", err)
	}
}

func TestServiceSetGetStarted(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	if err := svc.SetGetStarted(context.Background(), "GET_STARTED"); err != nil {
		t.Fatalf("SetGetStarted: %v", err)
	}
}

func TestServiceSetMenu(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	menu := json.RawMessage(`[{"locale":"default","call_to_actions":[]}]`)
	if err := svc.SetMenu(context.Background(), menu); err != nil {
		t.Fatalf("SetMenu: %v", err)
	}
}

func TestServiceSetIceBreakers(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	ib := json.RawMessage(`[{"question":"How can I help?","payload":"HELP"}]`)
	if err := svc.SetIceBreakers(context.Background(), ib); err != nil {
		t.Fatalf("SetIceBreakers: %v", err)
	}
}

func TestServiceDeleteProfileField(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	if err := svc.DeleteProfileField(context.Background(), "greeting"); err != nil {
		t.Fatalf("DeleteProfileField: %v", err)
	}
}

func TestServiceListConversations(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "conversations") {
			t.Errorf("expected path to contain conversations, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"id":            "conv_1",
					"updated_time":  "2026-01-01T00:00:00+0000",
					"message_count": 10,
					"participants":  map[string]any{"data": []map[string]any{{"name": "Alice"}, {"name": "Page"}}},
				},
			},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	list, err := svc.ListConversations(context.Background(), "111", 25)
	if err != nil {
		t.Fatalf("ListConversations: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(list))
	}
	if list[0].ID != "conv_1" {
		t.Errorf("expected conv_1, got %s", list[0].ID)
	}
	if list[0].Participants != "Alice, Page" {
		t.Errorf("expected 'Alice, Page', got %s", list[0].Participants)
	}
}

func TestServiceListConversationsError(t *testing.T) {
	srv, client := newTestGraphClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "bad", "code": 100},
		})
	})
	defer srv.Close()

	svc := messenger.NewService(client)
	_, err := svc.ListConversations(context.Background(), "111", 25)
	if err == nil {
		t.Error("expected error")
	}
}
